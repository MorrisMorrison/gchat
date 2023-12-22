package chatservice

import (
	"fmt"

	"github.com/MorrisMorrison/gchat/logger"
	"github.com/gorilla/websocket"
)

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
	ChatRoomNames = []string{"Lobby", "Room1", "Room2", "Room3", "Room4", "Room5"}
)

func InitializeChatRooms() {
	for _, chatRoomName := range ChatRoomNames {
		chatRooms[chatRoomName] = &ChatRoom{
			Name:  chatRoomName,
			Users: make([]*User, 0),
		}
	}
}

func FindUserChatRoom(user *User) (*ChatRoom, error) {
	for _, chatRoom := range chatRooms {
		for _, u := range chatRoom.Users {
			if user == u {
				return chatRoom, nil
			}
		}
	}
	return nil, fmt.Errorf("User %s not found in any chat room", user.Username)
}

func AddUserToChatRoom(chatRoomName string, user *User) {
	chatRooms[chatRoomName].Users = append(chatRooms[chatRoomName].Users, user)
}

func GetChatRoomUsersByChatRoomName(chatRoomName string) []*User {
	return chatRooms[chatRoomName].Users
}

func RemoveUserFromChatRoomByName(chatRoomName, username string) {
	chatRoom, exists := chatRooms[chatRoomName]
	if !exists {
		logger.Log.Debugf("ChatRoom %s not found", chatRoomName)
		return
	}

	user, exists := users[username]
	if !exists {
		logger.Log.Debugf("User %s not found", username)
	}

	RemoveUserFromChatRoomByReference(chatRoom, user)
}

func RemoveUserFromChatRoomByReference(chatRoom *ChatRoom, user *User) {
	for i, u := range chatRoom.Users {
		if user == u {
			chatRoom.Users[i] = chatRoom.Users[len(chatRoom.Users)-1]
			chatRoom.Users = chatRoom.Users[:len(chatRoom.Users)-1]

			logger.Log.Infof("User %s removed from chat room %s", user.Username, chatRoom.Name)
			return
		}
	}

	logger.Log.Infof("User %s not found in chat room %s", user.Username, chatRoom.Name)
}

func RemoveUserByName(username string) {
	delete(users, username)
}

func FindUserByName(username string) *User {
	user := users[username]
	return user
}

func AddUser(user *User) {
	users[user.Username] = user
}

func UserExists(username string) bool {
	if _, exists := users[username]; exists {
		return true
	}

	return false
}

func handleCloseConnection(user *User, _ int, _ string) error {
	logger.Log.Debugf("Close connection for user %s", user.Username)
	RemoveUserByName(user.Username)
	chatRoom, err := FindUserChatRoom(user)
	if err != nil {
		logger.Log.Errorf(err, "Could not find chat room for user %s", user.Username)
		return fmt.Errorf("could not find chat room for user")
	}
	RemoveUserFromChatRoomByReference(chatRoom, user)
	return nil
}

func (u *User) SetCloseHandler() {
	u.Conn.SetCloseHandler(func(code int, text string) error {
		return handleCloseConnection(u, code, text)
	})
}
