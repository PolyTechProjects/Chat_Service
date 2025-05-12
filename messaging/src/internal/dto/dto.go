package dto

import "github.com/google/uuid"

type MessageRequest struct {
	Body      string      `json:"body"`
	IsDirect  bool        `json:"is_direct"`
	CreatedAt uint64      `json:"created_at"`
	Files     []uuid.UUID `json:"files"`
}

type MessageResponse struct {
	MessageId     string   `json:"messageId"`
	SenderId      string   `json:"senderId"`
	DestinationId string   `json:"chatRoomId"`
	IsDirect      bool     `json:"is_direct"`
	Body          string   `json:"body"`
	IsEdited      bool     `json:"is_edited"`
	IsDeleted     bool     `json:"is_deleted"`
	CreatedAt     uint64   `json:"created_at"`
	Files         []string `json:"files"`
}

type HistoryResponse struct {
	Messages []MessageResponse `json:"messages"`
}

type NewMessageNotificationEvent struct {
	EventId       string `json:"event_id"`
	MessageId     string `json:"message_id"`
	SenderId      string `json:"sender_id"`
	DestinationId string `json:"destination_id"`
	ReceiverId    string `json:"receiver_id"`
	Body          string `json:"body"`
	IsDirect      bool   `json:"is_direct"`
	FilesCount    int    `json:"files_count"`
}
