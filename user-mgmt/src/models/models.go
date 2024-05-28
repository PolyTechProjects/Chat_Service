package models

import (
	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name        string    `gorm:"not null" json:"name"`
	Description string
	Avatar      string
}

func New(id uuid.UUID, name string) *User {
	return &User{Id: id, Name: name, Description: "", Avatar: ""}
}
