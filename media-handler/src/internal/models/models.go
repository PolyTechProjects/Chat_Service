package models

import (
	"github.com/google/uuid"
)

type Media struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	FileId string
}

func NewMedia(id uuid.UUID, fileId string) *Media {
	return &Media{ID: id, FileId: fileId}
}
