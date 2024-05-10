package models

import (
	"github.com/google/uuid"
)

type Media struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	FileId string
}

func New(id uuid.UUID, fileId string) *Media {
	return &Media{ID: id, FileId: fileId}
}

type SeaweedFSAssignResponse struct {
	Count     int    `json:"count"`
	Fid       string `json:"fid"`
	Url       string `json:"url"`
	PublicUrl string `json:"publicUrl"`
}

type SeaweedFSLookupResponse struct {
	VolumeId  int             `json:"volumeId"`
	Locations []PublicUrlXUrl `json:"locations"`
}

type PublicUrlXUrl struct {
	PublicUrl string `json:"publicUrl"`
	Url       string `json:"url"`
}

type UploadMediaRequest struct {
	MessageId string `json:"messageId"`
}

type MessageIdXFileId struct {
	MessageId uuid.UUID `json:"messageId"`
	FileId    uuid.UUID `json:"fileId"`
}
