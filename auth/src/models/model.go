package models

import (
	"log/slog"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id         uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Login      string    `gorm:"unique;not null;check:login <> ''"`
	Name       string    `gorm:"not null;check:name <> ''"`
	Pass       string    `gorm:"not null;check:pass <> ''"`
	KeycloakId string    `gorm:"unique;not null;check:keycloak_id <> ''"`
	Firstname  string    `gorm:"not null;check:firstname <> ''"`
	Lastname   string    `gorm:"not null;check:lastname <> ''"`
}

func New(login string, name string, pass string, firstname string, lastname string) (*User, error) {
	/*
		phonenumber, err := phonenumbers.Parse(login, "RU")
		if err != nil {
			return nil, err
		}
		if !phonenumbers.IsValidNumber(phonenumber) {
			err := errors.New("invalid phone number")
			return nil, err
		}
	*/
	_, err := mail.ParseAddress(login)
	if err != nil {
		slog.Error("Invalid email: " + err.Error())
		return nil, err
	}

	user := User{
		Id:        uuid.New(),
		Login:     login,
		Name:      name,
		Pass:      pass,
		Firstname: firstname,
		Lastname:  lastname,
	}

	return &user, nil
}

type RefreshToken struct {
	Id        uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	UserId    uuid.UUID `gorm:"unique;not null"`
	Value     string    `gorm:"not null;check:value <> ''"`
	ExpiredAt time.Time `gorm:"not null"`
}

func NewRefreshToken(userId uuid.UUID, value string) *RefreshToken {
	refreshToken := RefreshToken{
		Id:        uuid.New(),
		UserId:    userId,
		Value:     value,
		ExpiredAt: time.Now().Add(time.Hour * 24 * 30),
	}
	return &refreshToken
}
