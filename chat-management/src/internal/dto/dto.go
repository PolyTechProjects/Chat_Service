package dto

import "github.com/google/uuid"

type CreateChatRequest struct {
	Name        string
	Description string
	CreatorID   string
}

type UpdateChatRequest struct {
	ChatID      uuid.UUID
	Name        string
	Description string
}

type UserChatRequest struct {
	ChatID uuid.UUID
	UserID uuid.UUID
}

type AdminRequest struct {
	ChatID uuid.UUID
	UserID uuid.UUID
}

type ChatResponse struct {
	ChatID string
}
