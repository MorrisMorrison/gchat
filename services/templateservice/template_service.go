package templateservice

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/MorrisMorrison/gchat/services/chatservice"
	"github.com/MorrisMorrison/gchat/viewmodels"
)

func BuildChatRoomContentTemplate(chatRoomName string, username string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-room-content.html")
	if err != nil {
		fmt.Println("Error loading template chat-room-content.html")
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatservice.ChatRoomNames,
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room-content", viewModel)
	if err != nil {
		fmt.Println("Error parsing template chat-room-content.html")
	}

	return &buf
}

func BuildChatRoomTemplate(chatRoomName string, username string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-room.html", "templates/chat-room-content.html")
	if err != nil {
		fmt.Println("Error loading template chat-room.html")
	}

	viewModel := viewmodels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatservice.ChatRoomNames,
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room.html", viewModel)
	if err != nil {
		fmt.Println("Error parsing template chat-room.html")
	}

	return &buf
}

func BuildChatMessageTemplate(username string, message string) *bytes.Buffer {
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
