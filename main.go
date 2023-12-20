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
	fmt.Println(r.Body)
	fmt.Println(r.FormValue("username"))
	t, err := template.ParseFiles("templates/chat-room.html")
	if err != nil {
		fmt.Println("Error loading template chat-room.html")
		return
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     r.FormValue("username"),
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

func handleCloseConnection(code int, text string) error {
	return nil
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

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn.SetCloseHandler(handleCloseConnection)
	defer conn.Close()

	username := r.URL.Query().Get("username")

	users[username] = &User{
		Username: username,
		Conn:     conn,
	}

	chatRooms["Lobby"].Users = append(chatRooms["Lobby"].Users, users[username])

	fmt.Println(chatRooms)
	fmt.Println(chatRooms["Lobby"].Users)
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
			fmt.Println(chatMessage)
			fmt.Println(user)
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
