package dto

import (
	"example.com/chat-management/src/internal/models"
	"github.com/google/uuid"
)

type CreateChatRequest struct {
	Name        string
	Description string
	CreatorId   string
}

type UpdateChatRequest struct {
	ChatId      uuid.UUID
	Name        string
	Description string
}

type UserChatRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type AdminRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type ChatResponse struct {
	ChatId string
}

type GetChatRequest struct {
	ChatId uuid.UUID
}

type GetChatResponse struct {
	Chat   *models.Chat
	Users  []string
	Admins []string
}
