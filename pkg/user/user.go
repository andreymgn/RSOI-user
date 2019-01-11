package user

import (
	"context"
	"time"

	"github.com/go-redis/redis"

	pb "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/andreymgn/RSOI/services/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	AccessTokenExpirationTime  = time.Minute * 15
	RefreshTokenExpirationTime = time.Hour * 24 * 7 * 2
	OAuthCodeExpirationTime    = time.Minute
)

var (
	statusNotFound         = status.Error(codes.NotFound, "user not found")
	statusInvalidUUID      = status.Error(codes.InvalidArgument, "invalid UUID")
	statusInvalidToken     = status.Error(codes.Unauthenticated, "invalid grpc token")
	statusInvalidUserToken = status.Error(codes.Unauthenticated, "invalid user token")
	statusInvalidCode      = status.Error(codes.Unauthenticated, "invalid code")
)

func internalError(err error) error {
	return status.Error(codes.Internal, err.Error())
}

// UserInfo converts User to protobuf struct
func (u *User) UserInfo() *pb.UserInfo {
	result := new(pb.UserInfo)
	result.Uid = u.UID.String()
	result.Username = u.Username
	return result
}

// GetUserInfo returns User
func (s *Server) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.UserInfo, error) {
	uid, err := uuid.Parse(req.Uid)
	if err != nil {
		return nil, statusInvalidUUID
	}

	user, err := s.db.getUserInfo(uid)
	switch err {
	case nil:
		return user.UserInfo(), nil
	case errNotFound:
		return nil, statusNotFound
	default:
		return nil, internalError(err)
	}
}

// CreateUser creates a new user
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserInfo, error) {
	valid, err := s.checkServiceToken(req.Token)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	if len(req.Username) == 0 {
		return nil, status.Error(codes.InvalidArgument, "username is empty")
	}

	if len(req.Password) == 0 {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	user, err := s.db.create(req.Username, req.Password)
	if err != nil {
		return nil, internalError(err)
	}

	return user.UserInfo(), nil
}

// UpdateUser updates user
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	uid, err := uuid.Parse(req.Uid)
	if err != nil {
		return nil, statusInvalidUUID
	}

	err = s.db.update(uid, req.Password)
	switch err {
	case nil:
		return new(pb.UpdateUserResponse), nil
	case errNotFound:
		return nil, statusNotFound
	default:
		return nil, internalError(err)
	}
}

// DeleteUser deletes user
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	uid, err := uuid.Parse(req.Uid)
	if err != nil {
		return nil, statusInvalidUUID
	}

	err = s.db.delete(uid)
	switch err {
	case nil:
		return new(pb.DeleteUserResponse), nil
	case errNotFound:
		return nil, statusNotFound
	default:
		return nil, internalError(err)
	}
}

// GetServiceToken returns token for user service access
func (s *Server) GetServiceToken(ctx context.Context, req *pb.GetServiceTokenRequest) (*pb.GetServiceTokenResponse, error) {
	appID, appSecret := req.AppId, req.AppSecret
	token, err := s.apiTokenAuth.Add(appID, appSecret)
	switch err {
	case nil:
		res := new(pb.GetServiceTokenResponse)
		res.Token = token
		return res, nil
	case auth.ErrNotFound:
		return nil, statusNotFound
	case auth.ErrWrongSecret:
		return nil, status.Error(codes.Unauthenticated, "wrong secret")
	default:
		return nil, internalError(err)
	}
}

// GetAccessToken returns authorization token for user
func (s *Server) GetAccessToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetAccessTokenResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	uid, err := s.db.getUIDByUsername(req.Username)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	samePassword, err := s.db.checkPassword(uid, req.Password)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	if !samePassword {
		return nil, status.Error(codes.Unauthenticated, "wrong password")
	}

	token := uuid.New().String()
	err = s.accessTokenStorage.Set(token, uid.String(), AccessTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	res := new(pb.GetAccessTokenResponse)
	res.Token = token
	res.Uid = uid.String()
	return res, nil
}

// GetUserByAccessToken checks access token existance and refreshes token expiration time
func (s *Server) GetUserByAccessToken(ctx context.Context, req *pb.GetUserByAccessTokenRequest) (*pb.GetUserByAccessTokenResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	token := req.UserToken
	uid, err := s.accessTokenStorage.Get(token).Result()
	if err == redis.Nil {
		return nil, statusInvalidUserToken
	} else if err != nil {
		return nil, internalError(err)
	}

	if _, err := uuid.Parse(uid); err != nil {
		return nil, statusInvalidUserToken
	}

	err = s.accessTokenStorage.Expire(token, AccessTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	res := new(pb.GetUserByAccessTokenResponse)
	res.Uid = uid
	return res, nil
}

// GetRefreshToken returns token which can be used to refresh access token
func (s *Server) GetRefreshToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetRefreshTokenResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	uid, err := s.db.getUIDByUsername(req.Username)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	samePassword, err := s.db.checkPassword(uid, req.Password)
	if !samePassword {
		return nil, status.Error(codes.Unauthenticated, "wrong password")
	}

	token := uuid.New().String()
	err = s.refreshTokenStorage.Set(token, uid.String(), RefreshTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	res := new(pb.GetRefreshTokenResponse)
	res.Token = token
	return res, nil
}

// RefreshAccessToken returns new access and refresh tokens for user
func (s *Server) RefreshAccessToken(ctx context.Context, req *pb.RefreshAccessTokenRequest) (*pb.RefreshAccessTokenResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	token := req.RefreshToken
	uid, err := s.refreshTokenStorage.Get(token).Result()
	if err == redis.Nil {
		return nil, statusInvalidUserToken
	} else if err != nil {
		return nil, internalError(err)
	}

	if _, err := uuid.Parse(uid); err != nil {
		return nil, statusInvalidUserToken
	}

	err = s.refreshTokenStorage.Del(token).Err()
	if err != nil {
		return nil, internalError(err)
	}

	refreshToken := uuid.New().String()
	err = s.refreshTokenStorage.Set(refreshToken, uid, RefreshTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	accessToken := uuid.New().String()
	err = s.accessTokenStorage.Set(accessToken, uid, AccessTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	res := new(pb.RefreshAccessTokenResponse)
	res.RefreshToken = refreshToken
	res.AccessToken = accessToken
	return res, nil
}

// CreateApp creates new third-party app
func (s *Server) CreateApp(ctx context.Context, req *pb.CreateAppRequest) (*pb.CreateAppResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	owner, err := uuid.Parse(req.Owner)
	if err != nil {
		return nil, statusInvalidUUID
	}

	app, err := s.db.createApp(owner, req.Name)
	if err != nil {
		return nil, internalError(err)
	}

	resp := new(pb.CreateAppResponse)
	resp.Id = app.UID.String()
	resp.Secret = app.Secret.String()

	return resp, nil
}

// GetAppInfo returns public app information
func (s *Server) GetAppInfo(ctx context.Context, req *pb.GetAppInfoRequest) (*pb.GetAppInfoResponse, error) {
	appID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, statusInvalidUUID
	}

	appInfo, err := s.db.getAppInfo(appID)
	switch err {
	case nil:
		resp := new(pb.GetAppInfoResponse)
		resp.Name = appInfo.Name
		resp.Owner = appInfo.Owner.String()
		return resp, nil
	case errNotFound:
		return nil, statusNotFound
	default:
		return nil, internalError(err)
	}
}

// GetOAuthCode returns new oauth code
func (s *Server) GetOAuthCode(ctx context.Context, req *pb.GetOAuthCodeRequest) (*pb.GetOAuthCodeResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	uid, err := s.db.getUIDByUsername(req.Username)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	samePassword, err := s.db.checkPassword(uid, req.Password)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	if !samePassword {
		return nil, status.Error(codes.Unauthenticated, "wrong password")
	}

	code := uuid.New().String()

	err = s.oauthCodeStorage.Set(req.AppUid+code, uid.String(), time.Minute).Err()
	if err != nil {
		return nil, internalError(err)
	}

	resp := new(pb.GetOAuthCodeResponse)
	resp.Code = code

	return resp, nil
}

// GetTokenFromCode returns access and refresh tokens for user by oauth code
func (s *Server) GetTokenFromCode(ctx context.Context, req *pb.GetTokenFromCodeRequest) (*pb.GetTokenFromCodeResponse, error) {
	valid, err := s.checkServiceToken(req.ApiToken)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, statusInvalidToken
	}

	appUID, err := uuid.Parse(req.AppUid)
	if err != nil {
		return nil, statusInvalidUUID
	}

	appSecret, err := uuid.Parse(req.AppSecret)
	if err != nil {
		return nil, statusInvalidUUID
	}

	valid, err = s.db.isValidAppCredentials(appUID, appSecret)
	if err == errNotFound {
		return nil, statusNotFound
	} else if err != nil {
		return nil, internalError(err)
	}

	if !valid {
		return nil, status.Error(codes.Unauthenticated, "wrong appid appsecret pair")
	}

	uid, err := s.oauthCodeStorage.Get(req.AppUid + req.Code).Result()
	if err == redis.Nil {
		return nil, statusInvalidUserToken
	} else if err != nil {
		return nil, internalError(err)
	}

	err = s.oauthCodeStorage.Del(req.AppUid + req.Code).Err()
	if err != nil && err != redis.Nil {
		return nil, internalError(err)
	}

	accessToken := uuid.New().String()
	err = s.accessTokenStorage.Set(accessToken, uid, AccessTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	refreshToken := uuid.New().String()
	err = s.refreshTokenStorage.Set(refreshToken, uid, RefreshTokenExpirationTime).Err()
	if err != nil {
		return nil, internalError(err)
	}

	resp := new(pb.GetTokenFromCodeResponse)
	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return resp, nil
}
