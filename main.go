package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"websocket/config"
	"websocket/helpers"
	"websocket/redis"
	"websocket/router"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("DOCKER_ENV") != "true" {
		_ = godotenv.Load()
	}

	config.LoadEnv()
	r := router.SetupRouter()
	helpers.RedisClient = redis.InitRedis()

	port := os.Getenv("PORT_SOCKET")
	ip := os.Getenv("IP")

	fmt.Println("Server is running on port:", port)

	ln, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		log.Fatal("Socket listen error:", err)
	}

	if err := http.Serve(ln, r); err != nil {
		log.Fatal("Serve error:", err)
	}
}
