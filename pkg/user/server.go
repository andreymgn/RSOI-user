package user

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/andreymgn/RSOI/services/auth"
	"github.com/go-redis/redis"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// Server implements posts service
type Server struct {
	db                  datastore
	apiTokenAuth        auth.Auth
	accessTokenStorage  *redis.Client
	refreshTokenStorage *redis.Client
	oauthCodeStorage    *redis.Client
}

// NewServer returns a new server
func NewServer(connString, redisAddr, redisPassword string, apiTokenDBNum int, apiTokenknownApps map[string]string) (*Server, error) {
	db, err := newDB(connString)
	if err != nil {
		return nil, err
	}

	tokenStorage, err := auth.NewInternalAPITokenStorage(redisAddr, redisPassword, apiTokenDBNum, apiTokenknownApps)
	if err != nil {
		return nil, err
	}

	accessTokenStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum + 1,
	})

	_, err = accessTokenStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	refreshTokenStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum + 2,
	})

	_, err = refreshTokenStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	oauthCodeStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum + 3,
	})

	_, err = oauthCodeStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Server{db, tokenStorage, accessTokenStorage, refreshTokenStorage, oauthCodeStorage}, nil
}

// Start starts a server
func (s *Server) Start(port int, tracer opentracing.Tracer) error {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
	)
	pb.RegisterUserServer(server, s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return server.Serve(lis)
}

func (s *Server) checkServiceToken(token string) (bool, error) {
	exists, err := s.apiTokenAuth.Exists(token)
	if err != nil {
		return false, status.Error(codes.Internal, "api token auth error")
	}

	return exists, nil
}
