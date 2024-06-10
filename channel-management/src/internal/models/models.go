package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Channel struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Name        string
	Description string
	CreatorId   string
}

type UserChannel struct {
	gorm.Model
	ChannelId uuid.UUID `gorm:"type:uuid"`
	UserId    uuid.UUID `gorm:"type:uuid"`
}

type Admin struct {
	gorm.Model
	ChannelId uuid.UUID `gorm:"type:uuid"`
	UserId    uuid.UUID `gorm:"type:uuid"`
}
