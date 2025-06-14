package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"websocket/model"

	"github.com/gorilla/websocket"
)

var roomStore sync.Map // map[string]map[*websocket.Conn]int

func GenerateChatRoomID(postOwnerID uint, senderID uint) string {
	return "chat:postOwner:" + strconv.Itoa(int(postOwnerID)) + ":user:" + strconv.Itoa(int(senderID))
}

func RegisterConnectionToRoom(userID uint, roomID string, conn *websocket.Conn) {
	//Lấy ra mảng của roomID, nếu chưa thì tạo
	val, _ := roomStore.LoadOrStore(roomID, &sync.Map{})

	//Gán conn vào mảng các conn của roomID
	conns := val.(*sync.Map)
	conns.Store(conn, userID)

	fmt.Println("---Tạo room: " + roomID)
}

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

func ReadMessageHandler(conn *websocket.Conn, senderID uint) {
	//Hàm hủy chạy sau khi hàm chính kết thúc
	defer func() {
		fmt.Printf("Đóng kết nối user %d\n", senderID)
		RemoveConnectionFromAllRooms(conn)
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

				roomID := ""

				if data.IsOwner {
					roomID = GenerateChatRoomID(senderID, data.UserID)
				} else {
					roomID = GenerateChatRoomID(data.UserID, senderID)
				}

				response := model.EventResponse{
					Event:  "send_message_response",
					Status: "success",
					Data: model.SendMessageDataResponse{
						RoomID:    roomID,
						Message:   data.Message,
						SenderID:  senderID,
						TimeStamp: time.Now(),
					},
				}

				sendMessageOther(roomID, senderID, response)
			}

		case "join_room":
			{
				var data model.JoinRoomDataRequest

				if err := json.Unmarshal(evt.Data, &data); err != nil {
					log.Println("join_room data error:", err)
					continue
				}

				roomID := ""

				if data.IsOwner {
					roomID = GenerateChatRoomID(senderID, data.UserID)
				} else {
					roomID = GenerateChatRoomID(data.UserID, senderID)
				}

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

				roomID := ""

				if data.IsOwner {
					roomID = GenerateChatRoomID(senderID, data.UserID)
				} else {
					roomID = GenerateChatRoomID(data.UserID, senderID)
				}

				response := model.EventResponse{
					Event:  "left_room_response",
					Status: "success",
					Data: model.LeftRoomDataResponse{
						RoomID:    roomID,
						TimeStamp: time.Now(),
					},
				}

				sendMessageMyself(roomID, senderID, response)

				RemoveConnectionFromAllRooms(conn)
			}
		default:
			log.Println("unknown event:", evt.Event)
		}
	}
}

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

func sendMessageOther(roomID string, senderID uint, response model.EventResponse) {
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

	isInRoom := false

	conns.Range(func(key, value any) bool {
		userID := value.(uint)

		if userID == senderID {
			isInRoom = true

			return true
		}
		return true
	})

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
	}

	fmt.Println("")
}
