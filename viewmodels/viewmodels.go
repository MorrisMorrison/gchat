package viewmodels

type ChatMessageViewModel struct {
	Username        string
	DateTime        string
	Message         string
	IsSystemMessage bool
}

type ChatRoomViewModel struct {
	Username     string
	ChatRoomName string
}
