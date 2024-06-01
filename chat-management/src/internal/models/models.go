package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Chat struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Name        string
	Description string
	CreatorID   string
}

type UserChat struct {
	gorm.Model
	ChatID uuid.UUID `gorm:"type:uuid"`
	UserID uuid.UUID `gorm:"type:uuid"`
}

type Admin struct {
	gorm.Model
	ChatID uuid.UUID `gorm:"type:uuid"`
	UserID uuid.UUID `gorm:"type:uuid"`
}
