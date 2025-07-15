package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"websocket/grpcpb"
	"websocket/helpers"
	"websocket/redis"
	"websocket/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// 1. Kết nối tới gRPC Server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Không thể kết nối: %v", err)
	}

	defer conn.Close()

	helpers.GRPCConn = grpcpb.NewMessageHandlerClient(conn)

	fmt.Println("👀 Start main")

	_ = godotenv.Load()

	gin.SetMode(os.Getenv("GIN_MODE"))

	fmt.Println("Chuẩn bị connect")

	helpers.RedisClient = redis.InitRedis()

	r := router.SetupRouter()

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
