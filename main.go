package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	http.HandleFunc("/ws", handleWebSocketConnection)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", http.StripPrefix("/", fs))

	fmt.Println("## gchat server is running on port 8080 ##")
	http.ListenAndServe(":8080", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Body)
	fmt.Println(r.FormValue("username"))
	w.Write([]byte(getChatRoomHtml(r.FormValue("username"))))
}

func parseHtmxMessage(b []byte) map[string]any {
	var result map[string]any
	json.Unmarshal(b, &result)
	return result
}

func getChatRoomHtml(username string) string {
	chatRoomHtml := `<div class="mt-4" hx-ws="connect:ws:localhost:8080/ws?username=%s">
      <div id="ws_room" hx-swap="innerHTML">
      </div>
      <form hx-ws="send:submit">
        <input class="input is-primary" type="text" name="ws_message" hx-trigger="changed" placeholder="Send a message ...">
      </form>
    </div>
  `

	return fmt.Sprintf(chatRoomHtml, username)
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()
	username := r.URL.Query().Get("username")
	currentTime := time.Now()
	currentTimeString := currentTime.Format("02.01.2006 - 15:04:05")
	fmt.Println("Client connected")
	content := ` 
  <div hx-swap-oob="beforeend:#ws_room">
  <div class="message is-link"> 
          <div class="message-header">
            <p class="">%s - %s</p>
          </div>
          <div class="message-body">
            %s
          </div>
        </div>
  </div>
    `

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		message := parseHtmxMessage(p)["ws_message"]

		if err := conn.WriteMessage(messageType, []byte(fmt.Sprintf(content, username, currentTimeString, message))); err != nil {
			fmt.Println(err)
			return
		}

	}
}
