package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Chat struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Name        string
	Description string
	CreatorId   uuid.UUID `gorm:"type:uuid"`
}

type UserChat struct {
	gorm.Model
	ChatId uuid.UUID `gorm:"type:uuid"`
	UserId uuid.UUID `gorm:"type:uuid"`
}

type Admin struct {
	gorm.Model
	ChatId uuid.UUID `gorm:"type:uuid"`
	UserId uuid.UUID `gorm:"type:uuid"`
}
