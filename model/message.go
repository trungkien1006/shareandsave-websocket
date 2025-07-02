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
	SenderID   uint      `json:"senderID"`
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

type SendTransactionDataResponse struct {
	InterestID uint `json:"interestID"`
	ReceiverID uint `json:"receiverID"`
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

type SendTransactionRequest struct {
	InterestID uint `json:"interestID"`
	ReceiverID uint `json:"receiverID"`
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

type NotiSend struct {
	ID           uint      `json:"id"`
	SenderID     uint      `json:"senderID"`
	SenderName   string    `json:"senderName"`
	ReceiverID   uint      `json:"receiverID"`
	ReceiverName string    `json:"receiverName"`
	Type         string    `json:"type"`
	TargetType   string    `json:"targetType"`
	TargetID     uint      `json:"targetID"`
	Content      string    `json:"content"`
	IsRead       bool      `json:"isRead"`
	CreatedAt    time.Time `json:"createdAt"`
}
