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

var roomChatOneOne sync.Map // map[string]map[*websocket.Conn]int
var roomChatNoti sync.Map
var roomNoti sync.Map

func GenerateChatRoomID(interestID uint) string {
	return "chat:interest:" + strconv.Itoa(int(interestID))
}

func GenerateChatNotiRoomID(senderID uint) string {
	return "chatNoti:user:" + strconv.Itoa(int(senderID))
}

func GenerateNotiRoomID(receiverID uint) string {
	return "noti:user:" + strconv.Itoa(int(receiverID))
}

// Hàm tạo room public chat noti cho mỗi user
func RegisterConnectionToRoomChatNoti(roomID string, conn *websocket.Conn) {
	//Lấy ra mảng của roomID, nếu chưa thì tạo
	_, loaded := roomChatNoti.LoadOrStore(roomID, conn)

	if loaded {
		fmt.Println("---Room đã tồn tại: " + roomID)
	} else {
		fmt.Println("---Tạo room thành công: " + roomID)
	}
}

// Hàm tạo room public noti cho mỗi user
func RegisterConnectionToRoomNoti(roomID string, conn *websocket.Conn) {
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
	roomChatNoti.Delete(roomID)
	fmt.Println("---Xóa room: " + roomID)
}

// Hàm xóa room public noti
func RemoveConnectionFromRoomNoti(roomID string) {
	roomNoti.Delete(roomID)
	fmt.Println("---Xóa room: " + roomID)
}

// Hàm tạo room chat 1-1
func RegisterConnectionToRoom(userID uint, roomID string, conn *websocket.Conn) {
	//Lấy ra mảng của roomID, nếu chưa thì tạo
	val, _ := roomChatOneOne.LoadOrStore(roomID, &sync.Map{})

	//Gán conn vào mảng các conn của roomID
	userIDs := val.(*sync.Map)

	_, exists := userIDs.Load(userID)
	if exists {
		fmt.Println("Kết nối đã tồn tại trong room: " + roomID)
		return
	}

	count := 0

	userIDs.Range(func(nil, value any) bool {
		count++
		return true
	})

	if count < 2 {
		userIDs.Store(userID, conn)

		fmt.Println("---Tạo room thành công: " + roomID)
		return
	}

	fmt.Println("---Không thể join room: " + strconv.Itoa(count))
}

// Hàm xóa user khỏi tát cả room chat, nếu không còn user -> xóa room chat
func RemoveConnectionFromAllRooms(userID uint) {
	//Lặp qua 1 mảng các room
	roomChatOneOne.Range(func(key, value any) bool {
		//Lấy ra mảng các kết nối thuộc roomID
		roomID := key.(string)
		userIDs := value.(*sync.Map)

		fmt.Println("---Rời phòng: " + roomID)

		// Xoá kết nối ra khỏi room
		userIDs.Delete(userID)

		// Kiểm tra nếu không còn ai thì xoá luôn room
		empty := true
		userIDs.Range(func(_, _ any) bool {
			fmt.Println("---Kiểm tra còn người: " + roomID)
			empty = false
			return false // Dừng sớm nếu có ít nhất 1 người
		})

		if empty {
			fmt.Println("---Xóa phòng: " + roomID)
			roomChatOneOne.Delete(roomID)
		}

		return true
	})
}

// Hàm xóa user khỏi 1 room chat, nếu không còn user -> xóa room chat
func RemoveConnectionFromRoom(roomIDRemove string, userID uint) {
	//Lặp qua 1 mảng các room
	roomChatOneOne.Range(func(key, value any) bool {
		//Lấy ra mảng các kết nối thuộc roomID
		roomID := key.(string)
		userIDs := value.(*sync.Map)

		if roomID == roomIDRemove {
			// Xoá kết nối ra khỏi room
			userIDs.Delete(userID)

			// Kiểm tra nếu không còn ai thì xoá luôn room
			empty := true
			userIDs.Range(func(_, _ any) bool {
				empty = false
				return false // Dừng sớm nếu có ít nhất 1 người
			})

			if empty {
				roomChatOneOne.Delete(roomID)
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

	RegisterConnectionToRoomChatNoti(roomID, conn)

	sendMessageChatNoti(roomID, response)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"keep_alive"}`))
	}
}

// Hàm xử lí chat 1-1
func ReadMessageHandler(conn *websocket.Conn, senderID uint) {
	//Hàm hủy chạy sau khi hàm chính kết thúc
	defer func() {
		fmt.Printf("Đóng kết nối user %d\n", senderID)
		RemoveConnectionFromAllRooms(senderID)
		conn.Close()
	}()

	//Chạy vòng lặp đọc dữ liệu từ kết nối FE
	for {
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

		fmt.Printf("")
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
				fmt.Println("----Send message receiverID:", data.UserID)

				if isSended {
					sendMessageMyself(roomID, senderID, response)

					roomChatNoti := GenerateChatNotiRoomID(data.UserID)

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

					sendMessageChatNoti(roomChatNoti, notiResponse)
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

				RemoveConnectionFromRoom(roomID, senderID)
			}
		case "send_transaction":
			{
				var data model.SendTransactionRequest

				if err := json.Unmarshal(evt.Data, &data); err != nil {
					log.Println("send transaction data error:", err)
					continue
				}

				roomID := GenerateChatRoomID(data.InterestID)

				response := model.EventResponse{
					Event:  "send_transaction_response",
					Status: "success",
				}

				_, isSended := sendMessageOther(roomID, senderID, response)

				if isSended {
					response.Status = "error"
					response.Error = "Gửi tin nhắn không thành công"

					// sendMessageMyself(roomID, senderID, response)
				} else {
					response.Data = model.SendTransactionDataResponse{
						InterestID: data.InterestID,
						ReceiverID: data.ReceiverID,
					}

					// sendMessageMyself(roomID, senderID, response)

					sendMessageNoti(roomID, response)
				}
			}
		default:
			log.Println("unknown event:", evt.Event)
		}
	}

	fmt.Printf("")
}

func JoinRoomNotiHandler(conn *websocket.Conn, receiverID uint) {
	// Join room
	roomID := GenerateNotiRoomID(receiverID)

	fmt.Printf("Kết nối room noti của user: %d\n", receiverID)

	//Hàm hủy chạy sau khi hàm chính kết thúc
	defer func() {
		fmt.Printf("Đóng kết nối user %d\n", receiverID)
		RemoveConnectionFromRoomNoti(roomID)
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

	RegisterConnectionToRoomNoti(roomID, conn)

	sendMessageNoti(roomID, response)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"keep_alive"}`))
	}
}

func SendNoti(ctx context.Context, notis []map[string]string) error {
	fmt.Println("")
	fmt.Println("-----Bắt đầu gửi batch noti-----")

	var domainNotis []model.NotiSend

	timeLayout := "2006-01-02T15:04:05.000Z07:00"

	for _, value := range notis {
		ID, _ := strconv.Atoi(value["ID"])
		senderID, _ := strconv.Atoi(value["senderID"])
		receiverID, _ := strconv.Atoi(value["receiverID"])
		targetID, _ := strconv.Atoi(value["targetID"])
		createdAt, _ := time.Parse(timeLayout, value["createdAt"])
		isRead, _ := strconv.Atoi(value["isRead"])

		domainNotis = append(domainNotis, model.NotiSend{
			ID:           uint(ID),
			SenderID:     uint(senderID),
			SenderName:   value["senderName"],
			ReceiverID:   uint(receiverID),
			ReceiverName: value["receiverName"],
			Type:         value["type"],
			TargetType:   value["targetType"],
			TargetID:     uint(targetID),
			Content:      value["content"],
			IsRead:       isRead != 0,
			CreatedAt:    createdAt,
		})
	}

	for _, value := range domainNotis {
		if value.ReceiverID != 0 {
			roomID := GenerateNotiRoomID(value.ReceiverID)

			fmt.Println("")
			fmt.Println("-----")
			fmt.Println("Thông báo đến: " + roomID)
			fmt.Println("Nội dung thông báo: " + value.Content)

			response := model.EventResponse{
				Event:  "receive_noti_response",
				Status: "success",
				Data:   value,
			}

			sendMessageNoti(roomID, response)

			fmt.Println("-----")
			fmt.Println("")
		} else {
			fmt.Println("")
			fmt.Println("-----")
			fmt.Println("Gửi thông báo đến toàn hệ thống")
			fmt.Println("Nội dung thông báo: " + value.Content)

			response := model.EventResponse{
				Event:  "receive_noti_response",
				Status: "success",
				Data:   value,
			}

			sendMessageNotiToAllRooms(response)

			fmt.Println("-----")
			fmt.Println("")
		}
	}

	fmt.Println("-----Kết thúc gửi batch noti-----")
	fmt.Println("")

	return nil
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

func sendMessageNotiToAllRooms(response model.EventResponse) {
	// Ép response thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	// Duyệt qua toàn bộ các room đang kết nối trong roomNoti
	roomNoti.Range(func(key, value interface{}) bool {
		roomID := key.(string)
		conn := value.(*websocket.Conn)

		fmt.Println("→ Gửi tin nhắn đến room:", roomID)

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("Gửi lỗi đến", roomID, ":", err)

			// Xử lý nếu connection đã đóng hoặc lỗi
			conn.Close()
			roomNoti.Delete(roomID)
		} else {
			fmt.Println("✓ Gửi thành công đến:", roomID)
		}

		return true // tiếp tục vòng lặp
	})

	fmt.Println("↪ Gửi xong thông báo đến tất cả roomNoti.")
}

// Hàm gửi tin nhắn đến thông báo của user khác
func sendMessageChatNoti(roomID string, response model.EventResponse) {
	//Kiểm tra room có tồn tại hay không
	val, ok := roomChatNoti.Load(roomID)
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
	val, ok := roomChatOneOne.Load(roomID)
	if !ok {
		fmt.Println("Room not found:", roomID)
		return
	}

	fmt.Println("---Gửi tin nhắn đến:", roomID)

	userIDs := val.(*sync.Map)

	//Ép thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	//Duyệt qua mảng các conn gửi tin nhắn đến người khác
	userIDs.Range(func(key, value any) bool {
		conn := value.(*websocket.Conn)
		userID := key.(uint)

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
	val, ok := roomChatOneOne.Load(roomID)
	if !ok {
		fmt.Println("Room not found:", roomID)
		return 0, false
	}

	fmt.Println("---Gửi tin nhắn đến:", roomID)

	userIDs := val.(*sync.Map)

	isInRoom := false
	numUser := 0

	userIDs.Range(func(key, value any) bool {
		userID := key.(uint)

		numUser++

		if userID == senderID {
			isInRoom = true
		}

		return true
	})

	//Ép thành JSON
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Marshal error:", err)
		return 0, false
	}

	if isInRoom {
		//Duyệt qua mảng các conn gửi tin nhắn đến người khác
		userIDs.Range(func(key, value any) bool {
			conn := value.(*websocket.Conn)
			userID := key.(uint)

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

// func StartPingAllRooms() {
// 	ticker := time.NewTicker(10 * time.Second)

// 	go func() {
// 		for {
// 			<-ticker.C
// 			log.Println("[PING] Sending ping to all connections...")

// 			// Ping roomChatOneOne
// 			roomChatOneOne.Range(func(roomID, usersMap any) bool {
// 				users := usersMap.(*sync.Map)
// 				users.Range(func(_, conn any) bool {
// 					c := conn.(*websocket.Conn)
// 					err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(15*time.Second))
// 					if err != nil {
// 						log.Printf("Error ping roomChatOneOne (%s): %v\n", roomID, err)
// 					}
// 					return true
// 				})
// 				return true
// 			})

// 			// Ping roomChatNoti
// 			roomChatNoti.Range(func(roomID, conn any) bool {
// 				c := conn.(*websocket.Conn)
// 				rID := roomID.(string)
// 				err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(15*time.Second))
// 				if err != nil {
// 					log.Printf("Error ping roomChatNoti (%s): %v\n", rID, err)
// 					RemoveConnectionFromRoomChatNoti(rID)
// 				}
// 				return true
// 			})

// 			// Ping roomNoti
// 			roomNoti.Range(func(roomID, conn any) bool {
// 				c := conn.(*websocket.Conn)
// 				rID := roomID.(string)
// 				err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(15*time.Second))
// 				if err != nil {
// 					log.Printf("Error ping roomNoti (%s): %v\n", rID, err)
// 					RemoveConnectionFromRoomNoti(rID)
// 				}
// 				return true
// 			})
// 		}
// 	}()
// }
