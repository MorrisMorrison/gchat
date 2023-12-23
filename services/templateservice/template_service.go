package templateservice

import (
	"bytes"
	"html/template"
	"time"

	"github.com/MorrisMorrison/gchat/logger"
	chatService "github.com/MorrisMorrison/gchat/services/chatservice"
	configService "github.com/MorrisMorrison/gchat/services/configurationservice"
	viewModels "github.com/MorrisMorrison/gchat/viewmodels"
)

func BuildLoginTemplate(errorMessage string) (*bytes.Buffer, error) {
	t, err := template.ParseFiles("templates/login.html")
	if err != nil {
		logger.Log.Error(err, "Error loading template login.html")
		return nil, err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, viewModels.LoginViewModel{ErrorMessage: errorMessage})
	if err != nil {
		logger.Log.Error(err, "Error parsing template login.html")
		return nil, err
	}

	return &buf, nil
}

func BuildChatRoomContentTemplate(chatRoomName string, username string) (*bytes.Buffer, error) {
	t, err := template.ParseFiles("templates/chat-room-content.html")
	if err != nil {
		logger.Log.Error(err, "Error loading template chat-room-content.html")
		return nil, err
	}

	viewModel := viewModels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatService.ChatRoomNames,
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room-content", viewModel)
	if err != nil {
		logger.Log.Error(err, "Error parsing template chat-room-content.html")
		return nil, err
	}

	return &buf, nil
}

func BuildChatRoomTemplate(chatRoomName string, username string) (*bytes.Buffer, error) {
	t, err := template.ParseFiles("templates/chat-room.html", "templates/chat-room-content.html")
	if err != nil {
		logger.Log.Error(err, "Error loading template chat-room.html")
		return nil, err
	}

	viewModel := viewModels.ChatRoomViewModel{
		Username:     username,
		ChatRoomName: chatRoomName,
		Rooms:        chatService.ChatRoomNames,
		BaseUrl:      configService.GetBaseUrl(),
		Port:         configService.GetPort(),
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "chat-room.html", viewModel)
	if err != nil {
		logger.Log.Error(err, "Error parsing template chat-room.html")
		return nil, err
	}

	return &buf, nil
}

func BuildChatMessageTemplate(username string, message string) (*bytes.Buffer, error) {
	t, err := template.ParseFiles("templates/chat-message.html")
	if err != nil {
		logger.Log.Error(err, "Error loading template chat-message.html")
		return nil, err
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
		logger.Log.Error(err, "Error parsing template chat-message.html")
		return nil, err
	}

	return &buf, nil
}
