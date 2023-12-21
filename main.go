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
	users         = make(map[string]*User)
	chatRooms     = make(map[string]*ChatRoom)
	chatRoomNames = []string{"Lobby", "Room1", "Room2", "Room3", "Room4", "Room5"}
)

func main() {
	fmt.Println("### start gchat server ###")
	http.HandleFunc("/login", login)
	http.HandleFunc("/join", join)
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", http.StripPrefix("/", fs))

	initializeChatRooms()

	fmt.Println("## gchat server is running on port 8080 ##")
	http.ListenAndServe(":8080", nil)
}

func initializeChatRooms() {
	for _, chatRoomName := range chatRoomNames {
		chatRooms[chatRoomName] = &ChatRoom{
			Name:  chatRoomName,
			Users: make([]*User, 0),
		}
	}
}

func join(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	room := r.FormValue("room")

	fmt.Printf("User %s join room %s", username, room)
	fmt.Println()
	// remove user from current room
	user := users[username]
	currentChatRoom, err := findUserChatRoom(user)
	if err != nil {
		fmt.Println("Could not find user in any chatroom")
		return
	}

	removeUserFromChatRoomByReference(currentChatRoom, user)
	addUserToChatRoom(room, user)

	w.Write(buildChatRoomContentTemplate(room, username).Bytes())
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

	}

	w.Write(buildChatRoomTemplate("Lobby", username).Bytes())
}

func parseHtmxMessage(b []byte) map[string]string {
	var result map[string]string
	json.Unmarshal(b, &result)
	return result
}

func handleCloseConnection(user *User, _ int, _ string) error {
	fmt.Println("Close connection for user " + user.Username)
	delete(users, user.Username)
	chatRoom, err := findUserChatRoom(user)
	if err != nil {
		fmt.Println("Could not find chat room for user")
		return fmt.Errorf("could not find chat room for user")
	}
	removeUserFromChatRoomByReference(chatRoom, user)
	return nil
}

func findUserChatRoom(user *User) (*ChatRoom, error) {
	for _, chatRoom := range chatRooms {
		for _, u := range chatRoom.Users {
			if user == u {
				return chatRoom, nil
			}
		}
	}
	return nil, fmt.Errorf("User %s not found in any chat room", user.Username)
}

func addUserToChatRoom(chatRoomName string, user *User) {
	chatRooms[chatRoomName].Users = append(chatRooms[chatRoomName].Users, user)
}

func removeUserFromChatRoomByName(chatRoomName, username string) {
	chatRoom, exists := chatRooms[chatRoomName]
	if !exists {
		fmt.Println("ChatRoom not found")
		return
	}

	user, exists := users[username]
	if !exists {
		fmt.Println("User not found")
	}

	removeUserFromChatRoomByReference(chatRoom, user)
}

func removeUserFromChatRoomByReference(chatRoom *ChatRoom, user *User) {
	for i, u := range chatRoom.Users {
		if user == u {
			chatRoom.Users[i] = chatRoom.Users[len(chatRoom.Users)-1]
			chatRoom.Users = chatRoom.Users[:len(chatRoom.Users)-1]

			fmt.Printf("User %s removed from chat room %s\n", user.Username, chatRoom.Name)
			fmt.Println()
			return
		}
	}

	fmt.Printf("User %s not found in chat room %s\n", user.Username, chatRoom.Name)
}

func buildChatRoomContentTemplate(chatRoomName string, username string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-room-content.html")
	if err != nil {
		fmt.Println("Error loading template chat-room-content.html")
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatRoomNames,
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room-content", viewModel)
	if err != nil {
		fmt.Println("Error parsing template chat-room-content.html")
	}

	return &buf
}

func buildChatRoomTemplate(chatRoomName string, username string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-room.html", "templates/chat-room-content.html")
	if err != nil {
		fmt.Println("Error loading template chat-room.html")
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatRoomNames,
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room.html", viewModel)
	if err != nil {
		fmt.Println("Error parsing template chat-room.html")
	}

	return &buf
}

func buildChatMessageTemplate(username string, message string) *bytes.Buffer {
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
		chatMessage := buildChatMessageTemplate(username, message)

		chatRoom, err := findUserChatRoom(user)
		if err != nil {
			fmt.Print("Could not find current chat room")
			return
		}

		for _, user := range chatRooms[chatRoom.Name].Users {
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
