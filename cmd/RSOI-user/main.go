package main

import (
	"flag"
	"fmt"
)

func main() {
	connString := flag.String("conn", "", "PostgreSQL connection string")
	portNum := flag.Int("port", -1, "Port on which service well listen")
	jaegerAddr := flag.String("jaeger-addr", "", "Jaeger address")
	redisAddr := flag.String("redis-addr", "", "Redis address")
	redisPass := flag.String("redis-pass", "", "Redis password")
	redisDB := flag.Int("redis-db", -1, "Redis DB")

	flag.Parse()

	port := *portNum
	conn := *connString
	ja := *jaegerAddr
	ra := *redisAddr
	rp := *redisPass
	rdb := *redisDB

	fmt.Printf("running user service on port %d\n", port)
	err := runUser(port, conn, ja, ra, rp, rdb)

	if err != nil {
		fmt.Printf("finished with error %v", err)
	}
}
