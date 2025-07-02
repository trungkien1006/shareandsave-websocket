package handler

import (
	"fmt"
	"net/http"
	"websocket/helpers"
	"websocket/socket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		fmt.Println("Yêu cầu từ Origin:", r.Header.Get("Origin")) // Debug origin
		return true
	}, // Chấp nhận mọi origin (có thể điều chỉnh)
}

func HandleChatOneOnOne(c *gin.Context) {
	// Nếu middleware đã dừng, kiểm tra và phản hồi tại đây
	if c.IsAborted() {
		authErr, exists := c.Get("auth_error")
		if exists {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, authErr)))
		} else {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(`{"error":"unauthorized"}`))
		}
		return
	}

	userID, err := helpers.GetUintFromContext(c, "userID")
	if err != nil {
		fmt.Println("Lỗi khi lấy id từ context:", err)
		return
	}

	token, err := helpers.GetStringFromContext(c, "Sec-WebSocket-Protocol")
	if err != nil {
		fmt.Println("Lỗi khi lấy token từ context:", err)
		return
	}

	fmt.Println("Nhận yêu cầu nâng cấp WebSocket...")

	upgrader.Subprotocols = []string{token}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Lỗi khi nâng cấp WebSocket:", err)
		return
	}

	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}

	fmt.Println("User ID:", userID)

	go socket.ReadMessageHandler(conn, uint(userID))
}

func HandleChatNoti(c *gin.Context) {
	// Nếu middleware đã dừng, kiểm tra và phản hồi tại đây
	if c.IsAborted() {
		authErr, exists := c.Get("auth_error")
		if exists {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, authErr)))
		} else {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(`{"error":"unauthorized"}`))
		}
		return
	}

	userID, err := helpers.GetUintFromContext(c, "userID")
	if err != nil {
		fmt.Println("Lỗi khi lấy id từ context:", err)
		return
	}

	token, err := helpers.GetStringFromContext(c, "Sec-WebSocket-Protocol")
	if err != nil {
		fmt.Println("Lỗi khi lấy token từ context:", err)
		return
	}

	fmt.Println("Nhận yêu cầu nâng cấp WebSocket...")

	upgrader.Subprotocols = []string{token}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Lỗi khi nâng cấp WebSocket:", err)
		return
	}

	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}

	fmt.Println("User ID:", userID)

	go socket.SendPublicMessageHandler(conn, uint(userID))
}

func HandlerNoti(c *gin.Context) {
	// Nếu middleware đã dừng, kiểm tra và phản hồi tại đây
	if c.IsAborted() {
		authErr, exists := c.Get("auth_error")
		if exists {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, authErr)))
		} else {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte(`{"error":"unauthorized"}`))
		}
		return
	}

	userID, err := helpers.GetUintFromContext(c, "userID")
	if err != nil {
		fmt.Println("Lỗi khi lấy id từ context:", err)
		return
	}

	token, err := helpers.GetStringFromContext(c, "Sec-WebSocket-Protocol")
	if err != nil {
		fmt.Println("Lỗi khi lấy token từ context:", err)
		return
	}

	fmt.Println("Nhận yêu cầu nâng cấp WebSocket...")

	upgrader.Subprotocols = []string{token}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Lỗi khi nâng cấp WebSocket:", err)
		return
	}

	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}

	fmt.Println("User ID:", userID)

	go socket.JoinRoomNotiHandler(conn, uint(userID))
}
