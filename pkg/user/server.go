package user

import (
	"fmt"
	"log"
	"net"

	pb "github.com/andreymgn/RSOI-user/pkg/user/proto"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
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
	fmt.Println(redisAddr, redisPassword)

	accessTokenStorage := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       apiTokenDBNum,
	})
	fmt.Println("redis_client")
	fmt.Println(accessTokenStorage)

	_, err = accessTokenStorage.Ping().Result()
	if err != nil {
		return nil, err
	}
	fmt.Println("ping")

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
func (s *Server) Start(port int) error {
	server := grpc.NewServer()
	pb.RegisterUserServer(server, s)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return server.Serve(lis)
}
