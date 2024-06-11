package models

import "github.com/google/uuid"

type UserIdXDeviceToken struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	UserId      uuid.UUID `gorm:"type:uuid;"`
	DeviceToken string
}

type ReadyMessage struct {
	Message      Message     `json:"message"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ReceiversIds []uuid.UUID `json:"receivers_ids"`
}

type Message struct {
	Id         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	SenderId   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	ChatRoomId uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Body       string
	CreatedAt  uint64
	WithMedia  int
}
