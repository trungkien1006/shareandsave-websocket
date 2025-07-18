package router

import (
	"context"
	"websocket/handler"
	"websocket/helpers"
	"websocket/middleware"
	"websocket/worker"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	ctx := context.Background()

	stream := "notistream"
	group := "notigroup"
	consumer := "worker-noti"

	//run chat worker
	streamConsumer := worker.NewStreamConsumer(helpers.RedisClient, stream, group, consumer)
	streamConsumer.CreateConsumerGroup()
	notiHandler := handler.NewNotiHandler(streamConsumer)

	go notiHandler.Run(ctx)

	// go socket.StartPingAllRooms()

	r.GET("/chat", middleware.AuthGuard(), handler.HandleChatOneOnOne)
	r.GET("/chat-noti", middleware.AuthGuard(), handler.HandleChatNoti)
	r.GET("/noti", middleware.AuthGuard(), handler.HandlerNoti)

	return r
}
