package main

import (
	"log"
	"os"
	"strconv"
)

func main() {
	conn := os.Getenv("CONN")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("PORT parse error")
		return
	}

	redisAddr := os.Getenv("REDIS-ADDR")
	redisPass := os.Getenv("REDIS-PASS")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS-DB"))
	if err != nil {
		log.Println("REDIS-DB parse error")
		return
	}

	jaegerAddr := os.Getenv("JAEGER-ADDR")

	log.Printf("running user service on port %d\n", port)
	err = runUser(port, conn, redisAddr, redisPass, redisDB, jaegerAddr)

	if err != nil {
		log.Printf("finished with error %v", err)
	}
}
