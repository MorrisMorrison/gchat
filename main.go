package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/MorrisMorrison/gchat/viewmodels"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type User struct {
	Conn     *websocket.Conn
	Username string
}

type ChatRoom struct {
	Name  string
	Users []*User
}

var (
	users     = make(map[string]*User)
	chatRooms = make(map[string]*ChatRoom)
)

func main() {
	fmt.Println("### start gchat server ###")
	http.HandleFunc("/login", login)
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", http.StripPrefix("/", fs))

	chatRooms["Lobby"] = &ChatRoom{
		Name:  "Lobby",
		Users: make([]*User, 0),
	}

	fmt.Println("## gchat server is running on port 8080 ##")
	http.ListenAndServe(":8080", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if _, exists := users[username]; exists {
		errorMessage := "Username is already taken."

		t, err := template.ParseFiles("templates/login.html")
		if err != nil {
			fmt.Println("Error loading template login.html")
		}

		err = t.Execute(w, viewmodels.LoginViewModel{ErrorMessage: errorMessage})
		if err != nil {
			fmt.Println("Error parsing template login.html")
		}

		return
	}

	t, err := template.ParseFiles("templates/chat-room.html")
	if err != nil {
		fmt.Println("Error loading template chat-room.html")

		return
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: "Lobby",
	}

	err = t.Execute(w, viewModel)
	if err != nil {
		fmt.Println("Error parsing template chat-room.html")
		return
	}
}

func parseHtmxMessage(b []byte) map[string]string {
	var result map[string]string
	json.Unmarshal(b, &result)
	return result
}

func handleCloseConnection(user *User, code int, text string) error {
	fmt.Println("Close connection for user " + user.Username)
	delete(users, user.Username)
	removeUserFromChatRoom("Lobby", user.Username)
	return nil
}

func removeUserFromChatRoom(chatRoomName, username string) {
	chatRoom, exists := chatRooms[chatRoomName]
	if !exists {
		fmt.Println("ChatRoom not found")
		return
	}

	for i, user := range chatRoom.Users {
		if user.Username == username {
			// moves last user to current user index and truncates the slice
			chatRoom.Users[i] = chatRoom.Users[len(chatRoom.Users)-1]
			chatRoom.Users = chatRoom.Users[:len(chatRoom.Users)-1]

			user.Conn.Close()

			fmt.Printf("User %s removed from chat room %s\n", username, chatRoomName)
			return
		}
	}

	fmt.Printf("User %s not found in chat room %s\n", username, chatRoomName)
}

func buildChatMessage(username string, message string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-message.html")
	if err != nil {
		fmt.Println("Error loading template chat-message.html")
	}

	currentTime := time.Now()
	currentTimeString := currentTime.Format("02.01.2006 - 15:04:05")
	viewModel := viewmodels.ChatMessageViewModel{
		Username:        username,
		DateTime:        currentTimeString,
		Message:         message,
		IsSystemMessage: false,
	}

	var buf bytes.Buffer
	parseErr := t.Execute(&buf, viewModel)
	if parseErr != nil {
		fmt.Println("Error parsing template chat-message.html")
	}

	return &buf
}

func (u *User) SetCloseHandler() {
	u.Conn.SetCloseHandler(func(code int, text string) error {
		return handleCloseConnection(u, code, text)
	})
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	username := r.URL.Query().Get("username")
	user := &User{
		Username: username,
		Conn:     conn,
	}

	user.SetCloseHandler()

	users[username] = user
	chatRooms["Lobby"].Users = append(chatRooms["Lobby"].Users, users[username])
	defer conn.Close()

	fmt.Println("Client connected")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		message := parseHtmxMessage(p)["ws_message"]
		chatMessage := buildChatMessage(username, message)

		for _, user := range chatRooms["Lobby"].Users {
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
