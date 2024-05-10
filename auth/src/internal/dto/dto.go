package dto

type MessageRequest struct {
	SenderId   uint64 `json:"senderId"`
	ChatRoomId uint64 `json:"chatRoomId"`
	Body       string `json:"body"`
	SendTime   uint64 `json:"sendTime"`
}
