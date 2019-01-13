package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	conn := os.Getenv("CONN")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("PORT parse error")
		return
	}

	redisAddr := os.Getenv("REDIS-ADDR")
	redisPass := os.Getenv("REDIS-PASS")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS-DB"))
	if err != nil {
		fmt.Println("REDIS-DB parse error")
		return
	}

	fmt.Printf("running post service on port %d\n", port)
	err = runUser(port, conn, redisAddr, redisPass, redisDB)

	if err != nil {
		fmt.Printf("finished with error %v", err)
	}
}
