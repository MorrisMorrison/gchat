package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/MorrisMorrison/gchat/logger"
	chatService "github.com/MorrisMorrison/gchat/services/chatservice"
	configService "github.com/MorrisMorrison/gchat/services/configurationservice"
	templateService "github.com/MorrisMorrison/gchat/services/templateservice"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	logger.Log.Info("start gchat server.")

	port := configService.GetPort()

	http.HandleFunc("/login", login)
	http.HandleFunc("/join", join)
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", http.StripPrefix("/", fs))

	chatService.InitializeChatRooms()

	logger.Log.Info("gchat server is running on port 8080 ##")
	http.ListenAndServe(":"+port, nil)
}

func join(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed. Use POST."))
		return
	}

	username := r.FormValue("username")
	room := r.FormValue("room")

	logger.Log.Infof("User %s join room %s", username, room)
	// remove user from current room
	user := chatService.FindUserByName(username)
	currentChatRoom, err := chatService.FindUserChatRoom(user)
	if err != nil {
		logger.Log.Infof("Could not find user %s in any chatroom", username)
		return
	}

	chatService.RemoveUserFromChatRoomByReference(currentChatRoom, user)
	chatService.AddUserToChatRoom(room, user)

	t, err := templateService.BuildChatRoomContentTemplate(room, username)
	if err != nil {
		return
	}

	w.Write(t.Bytes())
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		w.Header().Set("Allow", "POST, GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed. Use GET or POST."))
		return
	}

	if r.Method == http.MethodGet {
		t, err := templateService.BuildLoginTemplate("")
		if err != nil {
			return
		}

		w.Write(t.Bytes())
		return
	}

	username := r.FormValue("username")
	if chatService.UserExists(username) {
		errorMessage := "Username is already taken."

		t, err := templateService.BuildLoginTemplate(errorMessage)
		if err != nil {
			return
		}

		w.Write(t.Bytes())
		return
	}

	t, err := templateService.BuildChatRoomTemplate("Lobby", username)
	if err != nil {
		return
	}

	w.Write(t.Bytes())
}

func parseHtmxMessage(b []byte) map[string]string {
	var result map[string]string
	json.Unmarshal(b, &result)
	return result
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed. Use GET."))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error(err, "Could not upgrade connection to websocket connection.")
		return
	}

	username := r.URL.Query().Get("username")
	if sanitizedUsername := template.JSEscapeString(username); sanitizedUsername == "" {
		logger.Log.Warningf("Possible XSS found in username %s", username)
		username = "Idiot"
	}

	user := &chatService.User{
		Username: username,
		Conn:     conn,
	}

	user.SetCloseHandler()

	chatService.AddUser(user)
	chatService.AddUserToChatRoom("Lobby", user)

	defer conn.Close()

	logger.Log.Info("Client connected")

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			logger.Log.Error(err, "Could not read message from websocket connection.")
			return
		}

		message := parseHtmxMessage(p)["ws_message"]
		if sanitizedMessage := template.JSEscapeString(message); sanitizedMessage == "" {
			logger.Log.Warningf("Possible XSS found in message %s", message)
			return
		}

		chatMessageTemplate, err := templateService.BuildChatMessageTemplate(username, message)
		if err != nil {
			return
		}

		chatRoom, err := chatService.FindUserChatRoom(user)
		if err != nil {
			logger.Log.Errorf(err, "Could not find current chat room for user %s", username)
			return
		}

		for _, user := range chatService.GetChatRoomUsersByChatRoomName(chatRoom.Name) {
			if user.Conn != nil {
				if err := user.Conn.WriteMessage(websocket.TextMessage, chatMessageTemplate.Bytes()); err != nil {
					logger.Log.Errorf(err, "Could not write message to websocket connection of user %s", username)
					return
				}
			} else {
				logger.Log.Info("Found user with nil connection")
				return
			}
		}
	}
}
