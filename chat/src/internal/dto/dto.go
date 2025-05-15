package dto

import (
	"example.com/chat/src/internal/models"
	"github.com/google/uuid"
)

type CreateChatRequest struct {
	Name               string      `json:"name"`
	CreatorId          uuid.UUID   `json:"creator_id"`
	ParticipantsIds    []uuid.UUID `json:"participants_ids"`
	IsChannel          bool        `json:"is_channel"`
	IsClosed           bool        `json:"is_closed"`
	JoinLink           string      `json:"join_link"`
	ProfilePic         string      `json:"profile_pic"`
	Description        string      `json:"description"`
	DefaultPermissions []string    `json:"default_permissions"`
}

type EditChatRequest struct {
	ChatId      uuid.UUID `json:"chat_id"`
	Name        string    `json:"name"`
	JoinLink    string    `json:"join_link"`
	ProfilePic  string    `json:"profile_pic"`
	Description string    `json:"description"`
}

type AddUsersRequest struct {
	ChatId  uuid.UUID   `json:"chat_id"`
	UserIds []uuid.UUID `json:"users_ids"`
}

type DeleteUsersRequest struct {
	ChatId  uuid.UUID   `json:"chat_id"`
	UserIds []uuid.UUID `json:"users_ids"`
}

type CreateRoleRequest struct {
	ChatId      uuid.UUID `json:"chat_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Permissions []string  `json:"permissions"`
	BaseRoleId  uuid.UUID `json:"base_role_id"`
}

type UpdateRoleRequest struct {
	ChatId      uuid.UUID `json:"chat_id"`
	RoleId      uuid.UUID `json:"role_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Permissions []string  `json:"permissions"`
}

type SetRoleRequest struct {
	ChatId        uuid.UUID   `json:"chat_id"`
	TargetUserIds []uuid.UUID `json:"target_user_ids"`
	RoleId        uuid.UUID   `json:"role_id"`
}

type ChangeUserNicknameRequest struct {
	ChatId   uuid.UUID `json:"chat_id"`
	UserId   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
}

type CreateDirectChatRequest struct {
	UserId uuid.UUID `json:"user_id"`
}

type LeaveFromChatsRequest struct {
	Request []LeaveFromChatRequest `json:"request"`
}

type LeaveFromChatRequest struct {
	ChatId   uuid.UUID `json:"chat_id"`
	IsDirect bool      `json:"is_direct"`
}

type GetChatResponse struct {
	Chat         *models.Chat    `json:"chat"`
	Participants []*Participants `json:"participants"`
}

type RoleResponse struct {
	RoleId      uuid.UUID `json:"role_id"`
	ChatId      uuid.UUID `json:"chat_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Permissions []string  `json:"permissions"`
}

type RolesResponse struct {
	Roles []*RoleResponse `json:"roles"`
}

type DirectChatResponse struct {
	ChatId       uuid.UUID `json:"chat_id"`
	FirstUserId  uuid.UUID `json:"first_user_id"`
	SecondUserId uuid.UUID `json:"second_user_id"`
}

type ChatsResponse struct {
	Chats       []*models.Chat        `json:"chats"`
	DirectChats []*DirectChatResponse `json:"direct_chats"`
}

type ChatUserResponse struct {
	ChatId   uuid.UUID `json:"chat_id"`
	UserId   uuid.UUID `json:"user_id"`
	RoleId   uuid.UUID `json:"role_id"`
	Nickname string    `json:"nickname"`
}

type Participants struct {
	UserId   string `json:"user_id"`
	RoleId   string `json:"role_id"`
	Nickname string `json:"nickname"`
}

type NotificationEvent interface{}

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
