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
	SenderID  uint      `json:"senderID"`
	RoomID    string    `json:"roomID"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
}

type SendMessageNotiDataResponse struct {
	Type      string    `json:"type" example:"following, followedBy"`
	UserID    uint      `json:"userID"`
	TimeStamp time.Time `json:"timestamp"`
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
	IsOwner bool   `json:"isOwner"`
	UserID  uint   `json:"userID"`
	Message string `json:"message"`
}

type JoinRoomDataRequest struct {
	IsOwner bool `json:"isOwner"`
	UserID  uint `json:"userID"`
}

type LeftRoomDataRequest struct {
	IsOwner bool `json:"isOwner"`
	UserID  uint `json:"userID"`
}
