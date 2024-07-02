package dto

import (
	"example.com/channel-management/src/internal/models"
	"github.com/google/uuid"
)

type CreateChannelRequest struct {
	Name        string
	Description string
	CreatorId   uuid.UUID
}

type DeleteChannelRequest struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type UpdateChannelRequest struct {
	ChannelId   uuid.UUID
	Name        string
	Description string
	UserId      uuid.UUID
}

type AdminRequest struct {
	ChannelId        uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
}

type IsAdminRequest struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type JoinChannelRequest struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type LeaveChannelRequest struct {
	ChannelId uuid.UUID
	UserId    uuid.UUID
}

type InviteUserRequest struct {
	ChannelId        uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
}

type KickUserRequest struct {
	ChannelId        uuid.UUID
	UserId           uuid.UUID
	RequestingUserId uuid.UUID
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
