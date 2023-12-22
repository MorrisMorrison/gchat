package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	chatService "github.com/MorrisMorrison/gchat/services/chatservice"
	templateService "github.com/MorrisMorrison/gchat/services/templateservice"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	fmt.Println("### start gchat server ###")
	http.HandleFunc("/login", login)
	http.HandleFunc("/join", join)
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", http.StripPrefix("/", fs))

	chatService.InitializeChatRooms()

	fmt.Println("## gchat server is running on port 8080 ##")
	http.ListenAndServe(":8080", nil)
}

func join(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	room := r.FormValue("room")

	fmt.Printf("User %s join room %s", username, room)
	fmt.Println()
	// remove user from current room
	user := chatService.FindUserByName(username)
	currentChatRoom, err := chatService.FindUserChatRoom(user)
	if err != nil {
		fmt.Println("Could not find user in any chatroom")
		return
	}

	chatService.RemoveUserFromChatRoomByReference(currentChatRoom, user)
	chatService.AddUserToChatRoom(room, user)

	w.Write(templateService.BuildChatRoomContentTemplate(room, username).Bytes())
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if chatService.UserExists(username) {
		errorMessage := "Username is already taken."

		buf := templateService.BuildLoginTemplate(errorMessage)
		w.Write(buf.Bytes())
		return
	}

	w.Write(templateService.BuildChatRoomTemplate("Lobby", username).Bytes())
}

func parseHtmxMessage(b []byte) map[string]string {
	var result map[string]string
	json.Unmarshal(b, &result)
	return result
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	username := r.URL.Query().Get("username")
	user := &chatService.User{
		Username: username,
		Conn:     conn,
	}

	user.SetCloseHandler()

	chatService.AddUser(user)
	chatService.AddUserToChatRoom("Lobby", user)

	defer conn.Close()

	fmt.Println("Client connected")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		message := parseHtmxMessage(p)["ws_message"]
		chatMessage := templateService.BuildChatMessageTemplate(username, message)

		chatRoom, err := chatService.FindUserChatRoom(user)
		if err != nil {
			fmt.Print("Could not find current chat room")
			return
		}

		for _, user := range chatService.GetChatRoomUsersByChatRoomName(chatRoom.Name) {
			if user.Conn != nil {
				if err := user.Conn.WriteMessage(websocket.TextMessage, chatMessage.Bytes()); err != nil {
					fmt.Println(err)
					return
				}
			} else {
				fmt.Println("Found user with nil connection")
				return
			}
		}
	}
}
