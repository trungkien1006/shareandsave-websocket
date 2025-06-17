package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"websocket/helpers"
	"websocket/model"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var roomStore sync.Map // map[string]map[*websocket.Conn]int
var roomNoti sync.Map

func GenerateChatRoomID(interestID uint) string {
	return "chat:interest:" + strconv.Itoa(int(interestID))
}

func GenerateChatNotiRoomID(senderID uint) string {
	return "chatNoti:user:" + strconv.Itoa(int(senderID))
}

// Hàm tạo room public chat noti cho mỗi user
func RegisterConnectionToRoomChatNoti(roomID string, conn *websocket.Conn) {
	//Lấy ra mảng của roomID, nếu chưa thì tạo
	_, loaded := roomNoti.LoadOrStore(roomID, conn)

	if loaded {
		fmt.Println("---Room đã tồn tại: " + roomID)
	} else {
		fmt.Println("---Tạo room thành công: " + roomID)
	}
}

// Hàm xóa room public chat noti
func RemoveConnectionFromRoomChatNoti(roomID string) {
	roomNoti.Delete(roomID)
	fmt.Println("---Xóa room: " + roomID)
}

// Hàm tạo room chat 1-1
func RegisterConnectionToRoom(userID uint, roomID string, conn *websocket.Conn) {
	//Lấy ra mảng của roomID, nếu chưa thì tạo
	val, _ := roomStore.LoadOrStore(roomID, &sync.Map{})

	//Gán conn vào mảng các conn của roomID
	conns := val.(*sync.Map)

	count := 0

	conns.Range(func(nil, value any) bool {
		count++
		return true
	})

	if count < 2 {
		conns.Store(conn, userID)

		fmt.Println("---Tạo room thành công: " + roomID)
		return
	}

	fmt.Println("---Không thể join room: " + strconv.Itoa(count))
}

// Hàm xóa user khỏi tát cả room chat, nếu không còn user -> xóa room chat
func RemoveConnectionFromAllRooms(conn *websocket.Conn) {
	//Lặp qua 1 mảng các room
	roomStore.Range(func(key, value any) bool {
		//Lấy ra mảng các kết nối thuộc roomID
		roomID := key.(string)
		conns := value.(*sync.Map)

		// Xoá kết nối ra khỏi room
		conns.Delete(conn)

		// Kiểm tra nếu không còn ai thì xoá luôn room
		empty := true
		conns.Range(func(_, _ any) bool {
			empty = false
			return false // Dừng sớm nếu có ít nhất 1 người
		})

		if empty {
			roomStore.Delete(roomID)
		}

		return true
	})
}

// Hàm xóa user khỏi 1 room chat, nếu không còn user -> xóa room chat
func RemoveConnectionFromRoom(roomIDRemove string, conn *websocket.Conn) {
	//Lặp qua 1 mảng các room
	roomStore.Range(func(key, value any) bool {
		//Lấy ra mảng các kết nối thuộc roomID
		roomID := key.(string)
		conns := value.(*sync.Map)

		if roomID == roomIDRemove {
			// Xoá kết nối ra khỏi room
			conns.Delete(conn)

			// Kiểm tra nếu không còn ai thì xoá luôn room
			empty := true
			conns.Range(func(_, _ any) bool {
				empty = false
				return false // Dừng sớm nếu có ít nhất 1 người
			})

			if empty {
				roomStore.Delete(roomID)
			}
		}

		return true
	})
}

// Hàm xử lí chat noti
func SendPublicMessageHandler(conn *websocket.Conn, senderID uint) {
	// Join room
	roomID := GenerateChatNotiRoomID(senderID)

	//Hàm hủy chạy sau khi hàm chính kết thúc
	defer func() {
		fmt.Printf("Đóng kết nối user %d\n", senderID)
		RemoveConnectionFromRoomChatNoti(roomID)
		conn.Close()
	}()

	response := model.EventResponse{
		Event:  "join_noti_room_response",
		Status: "success",
		Data: model.JoinRoomDataResponse{
			RoomID:    roomID,
			TimeStamp: time.Now(),
		},
	}

	for {
		RegisterConnectionToRoomChatNoti(roomID, conn)

		sendMessageNoti(roomID, response)

		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
	}
}

// Hàm xử lí chat 1-1
func ReadMessageHandler(conn *websocket.Conn, senderID uint) {
	//Hàm hủy chạy sau khi hàm chính kết thúc
	defer func() {
		fmt.Printf("Đóng kết nối user %d\n", senderID)
		RemoveConnectionFromAllRooms(conn)
		conn.Close()
	}()

	//Chạy vòng lặp đọc dữ liệu từ kết nối FE
	for {
		fmt.Printf("")

		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		//Giải mã dữ liệu nhận được
		var evt model.EventRequest
		if err := json.Unmarshal(msgBytes, &evt); err != nil {
			log.Println("event unmarshal error:", err)
			continue
		}

		fmt.Println("---Xử lí sự kiện:", evt.Event)

		//Xử lí sự kiện tương ứng
		switch evt.Event {
		case "send_message":
			{
				var data model.SendMessageDataRequest

				if err := json.Unmarshal(evt.Data, &data); err != nil {
					log.Println("send_message data error:", err)
					continue
				}

				roomID := GenerateChatRoomID(data.InterestID)

				if senderID == data.UserID {
					response := model.EventResponse{
						Event:  "send_message_response",
						Status: "error",
						Data:   nil,
						Error:  "Id người gửi và nhận trùng nhau",
					}

					sendMessageMyself(roomID, senderID, response)

					continue
				}

				response := model.EventResponse{
					Event:  "send_message_response",
					Status: "success",
					Data: model.SendMessageDataResponse{
						InterestID: data.InterestID,
						RoomID:     roomID,
						Message:    data.Message,
						SenderID:   senderID,
						TimeStamp:  time.Now(),
					},
				}

				numUser, isSended := sendMessageOther(roomID, senderID, response)

				fmt.Println("----Send message status:", isSended)

				if isSended {
					sendMessageMyself(roomID, senderID, response)

					roomNoti := GenerateChatNotiRoomID(data.UserID)

					notiType := ""

					if data.IsOwner {
						notiType = "following"
					} else {
						notiType = "followedBy"
					}

					notiResponse := model.EventResponse{
						Event:  "send_message_response",
						Status: "success",
						Data: model.SendMessageNotiDataResponse{
							InterestID: data.InterestID,
							Type:       notiType,
							SenderID:   senderID,
							TimeStamp:  time.Now(),
						},
					}

					isRead := 0

					if numUser == 2 {
						isRead = 1
					}

					redisMessage := model.RedisMessageSend{
						InterestID: data.InterestID,
						SenderID:   senderID,
						ReceiverID: data.UserID,
						Content:    data.Message,
						IsRead:     isRead,
						CreatedAt:  time.Now(),
					}

					fmt.Println("---Số lượng user trong room: " + strconv.Itoa(numUser))

					sendMessageToRedis(redisMessage)

					sendMessageNoti(roomNoti, notiResponse)
				} else {
					response.Status = "error"
					response.Data = nil
					response.Error = "Gửi tin nhắn không thành công"

					sendMessageMyself(roomID, senderID, response)
				}
			}

		case "join_room":
			{
				var data model.JoinRoomDataRequest

				if err := json.Unmarshal(evt.Data, &data); err != nil {
					log.Println("join_room data error:", err)
					continue
				}

				roomID := GenerateChatRoomID(data.InterestID)

				response := model.EventResponse{
					Event:  "join_room_response",
					Status: "success",
					Data: model.JoinRoomDataResponse{
						RoomID:    roomID,
						TimeStamp: time.Now(),
					},
				}

				RegisterConnectionToRoom(senderID, roomID, conn)

				sendMessageMyself(roomID, senderID, response)
			}
		case "left_room":
			{
				var data model.JoinRoomDataRequest

				if err := json.Unmarshal(evt.Data, &data); err != nil {
					log.Println("left_room data error:", err)
					continue
				}

				roomID := GenerateChatRoomID(data.InterestID)

				response := model.EventResponse{
					Event:  "left_room_response",
					Status: "success",
					Data: model.LeftRoomDataResponse{
						RoomID:    roomID,
						TimeStamp: time.Now(),
					},
				}

				sendMessageMyself(roomID, senderID, response)

				RemoveConnectionFromRoom(roomID, conn)
			}
		default:
			log.Println("unknown event:", evt.Event)
		}
	}

	fmt.Printf("")
}

// Hàm gửi tin nhắn đến thông báo của user khác
func sendMessageNoti(roomID string, response model.EventResponse) {
	//Kiểm tra room có tồn tại hay không
	val, ok := roomNoti.Load(roomID)
	if !ok {
		fmt.Println("Room not found:", roomID)
		return
	}

	conn := val.(*websocket.Conn)

	fmt.Println("---Gửi tin nhắn thông báo đến:", roomID)

	//Ép thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	fmt.Println("Tìm thấy room:", roomID)
	fmt.Println("Gửi tin nhắn thông báo:", string(data))

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("Send error", ":", err)
	}

	fmt.Println("---Gửi tin nhắn thông báo xong---")
}

// Hàm gửi tin nhắn đến chính mình
func sendMessageMyself(roomID string, senderID uint, response model.EventResponse) {
	//Kiểm tra room có tồn tại hay không
	val, ok := roomStore.Load(roomID)
	if !ok {
		fmt.Println("Room not found:", roomID)
		return
	}

	fmt.Println("---Gửi tin nhắn đến:", roomID)

	conns := val.(*sync.Map)

	//Ép thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	//Duyệt qua mảng các conn gửi tin nhắn đến người khác
	conns.Range(func(key, value any) bool {
		conn := key.(*websocket.Conn)
		userID := value.(uint)

		if userID == senderID {
			fmt.Println("Tìm thấy room:", roomID)
			fmt.Println("Gửi tin nhắn:", string(data))

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("Send error to user", userID, ":", err)
			}
		}
		return true
	})

	fmt.Println("---Gửi tin nhắn xong---")
	fmt.Println("")
}

// Hàm gửi tin nhắn đến các connect khác
func sendMessageOther(roomID string, senderID uint, response model.EventResponse) (int, bool) {
	//Kiểm tra room có tồn tại hay không
	val, ok := roomStore.Load(roomID)
	if !ok {
		fmt.Println("Room not found:", roomID)
		return 0, false
	}

	fmt.Println("---Gửi tin nhắn đến:", roomID)

	conns := val.(*sync.Map)

	isInRoom := false
	numUser := 0

	conns.Range(func(key, value any) bool {
		userID := value.(uint)

		numUser++

		if userID == senderID {
			isInRoom = true
		}

		return true
	})

	if numUser == 2 {

	}

	//Ép thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return 0, false
	}

	if isInRoom {
		//Duyệt qua mảng các conn gửi tin nhắn đến người khác
		conns.Range(func(key, value any) bool {
			conn := key.(*websocket.Conn)
			userID := value.(uint)

			if userID != senderID {
				fmt.Println("Tìm thấy room:", roomID)
				fmt.Println("Gửi tin nhắn:", string(data))

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					fmt.Println("Send error to user", userID, ":", err)
				}
			}
			return true
		})

		fmt.Println("---Gửi tin nhắn xong---")
	} else {
		fmt.Println("---Bạn đã thoát phòng rồi---")

		return numUser, false
	}

	return numUser, true
}

func sendMessageToRedis(data model.RedisMessageSend) {
	ctx := context.Background()

	fmt.Println("---Gửi tin nhắn vào redis stream")

	helpers.RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: "chatstream",
		Values: map[string]interface{}{
			"interestID": data.InterestID,
			"senderID":   data.SenderID,
			"receiverID": data.ReceiverID,
			"content":    data.Content,
			"isRead":     data.IsRead,
			"createdAt":  data.CreatedAt,
		},
	})

	fmt.Println("---Gửi tin nhắn vào redis stream hoàn tất")
}
