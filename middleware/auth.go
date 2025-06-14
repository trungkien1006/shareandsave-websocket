package middleware

import (
	"fmt"
	"strconv"
	"websocket/helpers"

	"github.com/gin-gonic/gin"
)

func AuthGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt := c.GetHeader("Sec-WebSocket-Protocol")

		if err := helpers.CheckJWT(c.Request.Context(), jwt); err != nil {
			fmt.Println("Có lỗi khi auth:", err.Error())

			// Lưu lỗi vào context để handler sau xử lý
			c.Set("auth_error", err.Error())
			c.Abort() // vẫn cần abort để dừng các middleware sau đó
			return
		}

		JWTSubject := helpers.GetTokenSubject(jwt)

		fmt.Println("---User ID in Auth--- :" + strconv.Itoa(int(JWTSubject.Id)))

		c.Set("userID", JWTSubject.Id)
		c.Set("device", JWTSubject.Device)
		c.Set("Sec-WebSocket-Protocol", jwt)

		c.Next()
	}
}
