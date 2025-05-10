package models

import (
	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name        string    `gorm:"not null" json:"name"`
	Firstname   string    `gorm:"not null" json:"firstname"`
	Lastname    string    `gorm:"not null" json:"lastname"`
	ProfilePic  string    `gorm:"default:0;not null" json:"profile_pic"`
	ProfileLink string    `gorm:"unique;not null" json:"profile_link"`
	Description string
}

func New(id uuid.UUID, name string, firstname string, lastname string) *User {
	return &User{
		Id:          id,
		Name:        name,
		Firstname:   firstname,
		Lastname:    lastname,
		ProfilePic:  "0",
		ProfileLink: generateProfileLink(),
		Description: "",
	}
}

func generateProfileLink() string {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic("SnowflakeNewNode failed: " + err.Error())
	}
	return node.Generate().Base36()
}
