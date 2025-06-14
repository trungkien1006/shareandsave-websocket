package router

import (
	"websocket/handler"
	"websocket/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ws", middleware.AuthGuard(), handler.HandleWebSocket)

	return r
}
