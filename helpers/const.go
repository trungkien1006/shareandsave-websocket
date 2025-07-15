package helpers

import (
	"websocket/grpcpb"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var GRPCConn grpcpb.MessageHandlerClient
