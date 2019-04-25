package main

import (
	"github.com/andreymgn/RSOI-user/pkg/user"
	"github.com/andreymgn/RSOI/pkg/tracer"
)

func runUser(port int, connString, redisAddr, redisPassword string, redisDB int, jaegerAddr string) error {
	tracer, closer, err := tracer.NewTracer("user", jaegerAddr)
	if err != nil {
		return err
	}

	defer closer.Close()

	server, err := user.NewServer(connString, redisAddr, redisPassword, redisDB)
	if err != nil {
		return err
	}

	return server.Start(port, tracer)
}
