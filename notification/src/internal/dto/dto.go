package dto

import "github.com/google/uuid"

type NotificationEvent interface{}

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

type NewChatCreatedNotificationEvent struct {
	ChatId     uuid.UUID `json:"chat_id"`
	Name       string    `json:"name"`
	ReceiverId uuid.UUID `json:"receiver_id"`
}

type NewUserInChatNotificationEvent struct {
	ChatId     uuid.UUID `json:"chat_id"`
	UserId     uuid.UUID `json:"user_id"`
	ReceiverId uuid.UUID `json:"receiver_id"`
}

type NewSubscriptionNotificationEvent struct {
	ChatId uuid.UUID `json:"chat_id"`
	UserId uuid.UUID `json:"user_id"`
}

type SendEmailEvent struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
