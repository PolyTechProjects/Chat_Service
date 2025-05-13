package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Chat struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name        string    `gorm:"not null;check:name <> ''"`
	CreatorId   uuid.UUID `gorm:"type:uuid"`
	IsChannel   bool      `gorm:"not null;default:false"`
	IsClosed    bool      `gorm:"not null;default:true"`
	JoinLink    string    `gorm:"unique;default:NULL"`
	ProfilePic  string    `gorm:"default:0;not null;check:profile_pic <> ''"`
	Description string
}

type DirectChat struct {
	Id           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	FirstUserId  uuid.UUID `gorm:"type:uuid;not null;index:idx_direct_chat_users"`
	SecondUserId uuid.UUID `gorm:"type:uuid;not null;index:idx_direct_chat_users"`
}

type ChatUser struct {
	gorm.Model
	ChatId   uuid.UUID `gorm:"type:uuid"`
	UserId   uuid.UUID `gorm:"type:uuid"`
	RoleId   uuid.UUID `gorm:"type:uuid"`
	Nickname string    `gorm:"not null;check:nickname <> ''"`
}

// IsAdmin and IsDefault cannot be deleted
type Role struct {
	Id             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name           string    `gorm:"not null;check:name <> ''"`
	Description    string
	Color          string
	ChatId         uuid.UUID `gorm:"type:uuid"`
	IsAdmin        bool      `gorm:"not null;default:false"`
	IsDefault      bool      `gorm:"not null;default:false"`
	BasedOnDefault bool      `gorm:"not null;default:true"`
}

type RolePermission struct {
	Id         uint
	RoleId     uuid.UUID `gorm:"type:uuid"`
	Permission Permission
}

type Permission string

const (
	CAN_WRITE_MESSAGE          Permission = "CAN_WRITE_MESSAGE"
	CAN_PIN_MESSAGE            Permission = "CAN_PIN_MESSAGE"
	CAN_EDIT_OWN_MESSAGE       Permission = "CAN_EDIT_MESSAGE"
	CAN_EDIT_OTHERS_MESSAGE    Permission = "CAN_EDIT_OTHERS_MESSAGE"
	CAN_DELETE_OWN_MESSAGE     Permission = "CAN_DELETE_MESSAGE"
	CAN_DELETE_OTHERS_MESSAGE  Permission = "CAN_DELETE_OTHERS_MESSAGE"
	CAN_SEND_FILE              Permission = "CAN_SEND_FILE"
	CAN_CHANGE_OWN_NICKNAME    Permission = "CAN_CHANGE_OWN_NICKNAME"
	CAN_CHANGE_OTHERS_NICKNAME Permission = "CAN_CHANGE_OTHERS_NICKNAME"
	CAN_EDIT_CHAT              Permission = "CAN_EDIT_CHAT"
	CAN_DELETE_CHAT            Permission = "CAN_DELETE_CHAT"
	CAN_CREATE_ROLE            Permission = "CAN_CREATE_ROLE"
	CAN_EDIT_ROLE              Permission = "CAN_EDIT_ROLE"
	CAN_DELETE_ROLE            Permission = "CAN_DELETE_ROLE"
	CAN_SET_ROLES              Permission = "CAN_SET_ROLES"
	CAN_ADD_USERS              Permission = "CAN_ADD_USERS"
	CAN_DELETE_USERS           Permission = "CAN_DELETE_USERS"
	//CAN_REPOST_MESSAGE         Permission = "CAN_REPOST_MESSAGE"
)
