package dto

import (
	"example.com/channel-management/src/internal/models"
	"github.com/google/uuid"
)

type CreateChannelDTO struct {
	Name        string
	Description string
	CreatorId   uuid.UUID
}

type UpdateChannelDTO struct {
	Id          uuid.UUID
	Name        string
	Description string
}

type AddUserDTO struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type RemoveUserDTO struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type AdminDTO struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type GetChannelRequest struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type GetChannelResponse struct {
	Channel *models.Channel
	Users   []string
	Admins  []string
}
