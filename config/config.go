package config

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func LoadEnv() {
	gin.SetMode(os.Getenv("GIN_MODE"))
	if os.Getenv("PORT_SOCKET") == "" {
		log.Fatal("PORT_SOCKET not set in .env")
	}
}
