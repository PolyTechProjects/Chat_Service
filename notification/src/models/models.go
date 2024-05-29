package models

import "github.com/google/uuid"

type UserIdXDeviceToken struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	UserId      uuid.UUID `gorm:"type:uuid;"`
	DeviceToken string
}
