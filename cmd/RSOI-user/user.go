package main

import (
	"github.com/andreymgn/RSOI-user/pkg/user"
)

func runUser(port int, connString, redisAddr, redisPassword string, redisDB int) error {
	server, err := user.NewServer(connString, redisAddr, redisPassword, redisDB)
	if err != nil {
		return err
	}

	return server.Start(port)
}
