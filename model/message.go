package model

import (
	"encoding/json"
	"time"
)

// response
type EventResponse struct {
	Event  string      `json:"event"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"` // chưa giải mã vội
	Error  string      `json:"error"`
}

type SendMessageDataResponse struct {
	SenderID   uint      `json:"senderID"`
	InterestID uint      `json:"interestID"`
	RoomID     string    `json:"roomID"`
	Message    string    `json:"message"`
	TimeStamp  time.Time `json:"timestamp"`
}

type SendMessageNotiDataResponse struct {
	Type       string    `json:"type" example:"following, followedBy"`
	InterestID uint      `json:"interestID"`
	UserID     uint      `json:"userID"`
	TimeStamp  time.Time `json:"timestamp"`
}

type JoinRoomDataResponse struct {
	RoomID    string    `json:"roomID"`
	TimeStamp time.Time `json:"timestamp"`
}

type LeftRoomDataResponse struct {
	RoomID    string    `json:"roomID"`
	TimeStamp time.Time `json:"timestamp"`
}

// request
type EventRequest struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"` // chưa giải mã vội
}

type SendMessageDataRequest struct {
	InterestID uint   `json:"interestID"`
	IsOwner    bool   `json:"isOwner"`
	UserID     uint   `json:"userID"`
	Message    string `json:"message"`
}

type JoinRoomDataRequest struct {
	InterestID uint `json:"interestID"`
}

type LeftRoomDataRequest struct {
	InterestID uint `json:"interestID"`
}

// redis
type RedisMessageSend struct {
	InterestID uint
	SenderID   uint
	ReceiverID uint
	Content    string
	IsRead     int
	CreatedAt  time.Time
}
