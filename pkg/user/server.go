package user

import (
	"fmt"
	"net"

	pb "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/go-redis/redis"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server implements posts service
type Server struct {
	db                  datastore
	accessTokenStorage  *redis.Client
	refreshTokenStorage *redis.Client
	oauthCodeStorage    *redis.Client
}

// NewServer returns a new server
func NewServer(connString, redisAddr, redisPassword string, apiTokenDBNum int) (*Server, error) {
	db, err := newDB(connString)
	if err != nil {
		return nil, err
	}

	accessTokenStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum,
	})

	_, err = accessTokenStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	refreshTokenStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum + 1,
	})

	_, err = refreshTokenStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	oauthCodeStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum + 2,
	})

	_, err = oauthCodeStorage.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Server{db, accessTokenStorage, refreshTokenStorage, oauthCodeStorage}, nil
}

// Start starts a server
func (s *Server) Start(port int, tracer opentracing.Tracer) error {
	creds, err := credentials.NewServerTLSFromFile("/cert.pem", "/key.pem")
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
	)
	pb.RegisterUserServer(server, s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	return server.Serve(lis)
}
