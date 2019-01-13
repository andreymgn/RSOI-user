package main

import (
	"log"

	"github.com/andreymgn/RSOI-user/pkg/user"
)

const (
	UserAppID     = "UserAPI"
	UserAppSecret = "fzFKf3g6QeIdqbP7"
)

func runUser(port int, connString, redisAddr, redisPassword string, redisDB int) error {
	knownKeys := map[string]string{UserAppID: UserAppSecret}

	server, err := user.NewServer(connString, redisAddr, redisPassword, redisDB, knownKeys)
	if err != nil {
		log.Fatal(err)
	}

	return server.Start(port)
}
