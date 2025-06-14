package router

import (
	"websocket/handler"
	"websocket/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/chat", middleware.AuthGuard(), handler.HandleChatOneOnOne)
	r.GET("/chat-noti", middleware.AuthGuard(), handler.HandleChatNoti)

	return r
}
