package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/MorrisMorrison/gchat/services/chatservice"
	"github.com/MorrisMorrison/gchat/services/templateservice"
	"github.com/MorrisMorrison/gchat/viewmodels"
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

	chatservice.InitializeChatRooms()

	fmt.Println("## gchat server is running on port 8080 ##")
	http.ListenAndServe(":8080", nil)
}

func join(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	room := r.FormValue("room")

	fmt.Printf("User %s join room %s", username, room)
	fmt.Println()
	// remove user from current room
	user := chatservice.FindUserByName(username)
	currentChatRoom, err := chatservice.FindUserChatRoom(user)
	if err != nil {
		fmt.Println("Could not find user in any chatroom")
		return
	}

	chatservice.RemoveUserFromChatRoomByReference(currentChatRoom, user)
	chatservice.AddUserToChatRoom(room, user)

	w.Write(templateservice.BuildChatRoomContentTemplate(room, username).Bytes())
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if chatservice.UserExists(username) {
		errorMessage := "Username is already taken."

		t, err := template.ParseFiles("templates/login.html")
		if err != nil {
			fmt.Println("Error loading template login.html")
		}

		err = t.Execute(w, viewmodels.LoginViewModel{ErrorMessage: errorMessage})
		if err != nil {
			fmt.Println("Error parsing template login.html")
		}

	}

	w.Write(templateservice.BuildChatRoomTemplate("Lobby", username).Bytes())
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
	user := &chatservice.User{
		Username: username,
		Conn:     conn,
	}

	user.SetCloseHandler()

	chatservice.AddUser(user)
	chatservice.AddUserToChatRoom("Lobby", user)

	defer conn.Close()

	fmt.Println("Client connected")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		message := parseHtmxMessage(p)["ws_message"]
		chatMessage := templateservice.BuildChatMessageTemplate(username, message)

		chatRoom, err := chatservice.FindUserChatRoom(user)
		if err != nil {
			fmt.Print("Could not find current chat room")
			return
		}

		for _, user := range chatservice.GetChatRoomUsersByChatRoomName(chatRoom.Name) {
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
