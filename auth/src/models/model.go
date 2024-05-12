package models

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

type User struct {
	Id    uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Login string    `gorm:"unique;not null"`
	Name  string    `gorm:"not null"`
	Pass  string    `gorm:"not null"`
}

func New(login string, name string, pass string) (*User, error) {
	phonenumber, err := phonenumbers.Parse(login, "RU")
	if err != nil {
		return nil, err
	}
	if !phonenumbers.IsValidNumber(phonenumber) {
		err := errors.New("invalid phone number")
		return nil, err
	}

	user := User{
		Id:    uuid.New(),
		Login: login,
		Name:  name,
		Pass:  pass,
	}

	return &user, nil
}
