package templateservice

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	chatService "github.com/MorrisMorrison/gchat/services/chatservice"
	viewModels "github.com/MorrisMorrison/gchat/viewmodels"
)

func BuildLoginTemplate(errorMessage string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		fmt.Println("Error loading template login.html")
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, viewModels.LoginViewModel{ErrorMessage: errorMessage})
	if err != nil {
		fmt.Println("Error parsing template login.html")
	}

	return &buf
}

func BuildChatRoomContentTemplate(chatRoomName string, username string) *bytes.Buffer {
	t, err := template.ParseFiles("templates/chat-room-content.html")
	if err != nil {
		fmt.Println("Error loading template chat-room-content.html")
	}

	viewModel := viewModels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatService.ChatRoomNames,
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

	viewModel := viewModels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatService.ChatRoomNames,
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
	viewModel := viewModels.ChatMessageViewModel{
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
