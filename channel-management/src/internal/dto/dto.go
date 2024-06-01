package dto

import "github.com/google/uuid"

type CreateChannelDTO struct {
	Name        string
	Description string
	CreatorID   uuid.UUID
}

type UpdateChannelDTO struct {
	ID          uuid.UUID
	Name        string
	Description string
}

type AddUserDTO struct {
	ChannelID uuid.UUID
	UserID    uuid.UUID
}

type RemoveUserDTO struct {
	ChannelID uuid.UUID
	UserID    uuid.UUID
}

type AdminDTO struct {
	ChannelID uuid.UUID
	UserID    uuid.UUID
}
