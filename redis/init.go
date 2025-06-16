package redis

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func InitRedis() *redis.Client {
	once.Do(func() {
		redisHost := os.Getenv("REDIS_HOST")

		redisClient = redis.NewClient(&redis.Options{
			Addr:         redisHost + ":6379", // hoặc "redis:6379" nếu dùng Docker
			PoolSize:     20,
			MinIdleConns: 5,
		})

		// Kiểm tra kết nối
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		pong, err := redisClient.Ping(ctx).Result()
		if err != nil {
			panic(fmt.Sprintf("❌ Không thể kết nối Redis: %v", err))
		}

		fmt.Println("✅ Kết nối Redis thành công:", pong)
	})

	return redisClient
}
