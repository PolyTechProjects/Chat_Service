package dto

import (
	"example.com/chat-management/src/internal/models"
	"github.com/google/uuid"
)

type CreateChatRequest struct {
	Name        string
	Description string
	CreatorId   uuid.UUID
}

type DeleteChatRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type UpdateChatRequest struct {
	ChatId      uuid.UUID
	Name        string
	Description string
	UserId      uuid.UUID
}

type AdminRequest struct {
	ChatId           uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
}

type IsAdminRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type JoinChatRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type LeaveChatRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type InviteUserRequest struct {
	ChatId           uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
}

type KickUserRequest struct {
	ChatId           uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
}

type ChatResponse struct {
	ChatId string
}

type GetChatRequest struct {
	ChatId uuid.UUID
	UserId uuid.UUID
}

type GetChatResponse struct {
	Chat   *models.Chat
	Users  []string
	Admins []string
}
