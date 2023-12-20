package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type User struct {
	conn     *websocket.Conn
	username string
}

type ChatMessageViewModel struct {
	Username string
	DateTime string
	Message  string
}

type ChatRoomViewModel struct {
	Username string
}

var (
	usernameToConnection = make(map[string]*websocket.Conn)
	users                = make([]*User, 20)
	chatRooms            = make([]string, 20)
)

func main() {
	fmt.Println("### start gchat server ###")
	http.HandleFunc("/login", login)
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", http.StripPrefix("/", fs))

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

	viewModel := ChatRoomViewModel{
		Username: r.FormValue("username"),
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

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.SetCloseHandler(handleCloseConnection)
	defer conn.Close()
	username := r.URL.Query().Get("username")
	if _, exists := usernameToConnection[username]; exists {
		return
	}

	usernameToConnection[username] = conn

	fmt.Println("Client connected")

	t, err := template.ParseFiles("templates/chat-message.html")
	if err != nil {
		fmt.Println("Error loading template chat-message.html")
		return
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		currentTime := time.Now()
		currentTimeString := currentTime.Format("02.01.2006 - 15:04:05")
		message := parseHtmxMessage(p)["ws_message"]
		viewModel := ChatMessageViewModel{
			Username: username,
			DateTime: currentTimeString,
			Message:  message,
		}

		var buf bytes.Buffer
		parseErr := t.Execute(&buf, viewModel)
		if parseErr != nil {
			fmt.Println("Error parsing template chat-message.html")
		}
		// TODO broadcast to all connections
		if err := conn.WriteMessage(websocket.TextMessage, buf.Bytes()); err != nil {
			fmt.Println(err)
			return
		}

	}
}
