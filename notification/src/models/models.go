package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Notification struct {
	Id         uuid.UUID        `gorm:"primary_key"`
	ReceiverId uuid.UUID        `gorm:"type:uuid;not null"`
	Type       NotificationType `gorm:"not null"`
	Body       *json.RawMessage `gorm:"not null;type:jsonb"`
}

type Subscription struct {
	Id     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	ChatId uuid.UUID `gorm:"type:uuid;not null"`
	UserId uuid.UUID `gorm:"type:uuid;not null"`
}

type NotificationType string

const (
	PUSH NotificationType = "PUSH"
	MAIL NotificationType = "MAIL"
)
