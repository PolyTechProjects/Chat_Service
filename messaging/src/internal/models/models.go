package models

import (
	"strings"

	"example.com/messaging/src/internal/dto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Id            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	SenderId      uuid.UUID `gorm:"type:uuid;not null"`
	DestinationId uuid.UUID `gorm:"type:uuid;not null"`
	IsDirect      bool      `gorm:"not null;default:false"`
	Body          string    `gorm:"not null;default:false"`
	IsEdited      bool      `gorm:"not null;default:false"`
	IsDeleted     bool      `gorm:"not null;default:false"`
	CreatedAt     uint64    `gorm:"not null"`
	Files         string
	FilesCount    int
}

func MapRequestToMessage(req *dto.MessageRequest, senderId uuid.UUID, destinationId uuid.UUID) (*Message, error) {
	messageUUID := uuid.New()
	files := strings.Builder{}
	for _, file := range req.Files {
		_, err := files.WriteString(file.String() + ",")
		if err != nil {
			return nil, err
		}
	}
	return &Message{
		Id:            messageUUID,
		SenderId:      senderId,
		DestinationId: destinationId,
		IsDirect:      req.IsDirect,
		Body:          req.Body,
		IsEdited:      false,
		IsDeleted:     false,
		CreatedAt:     req.CreatedAt,
		Files:         files.String(),
		FilesCount:    len(req.Files),
	}, nil
}

func MapMessageToResponse(message *Message) *dto.MessageResponse {
	messageId := message.Id.String()
	senderId := message.SenderId.String()
	destinationId := message.DestinationId.String()
	files := make([]string, message.FilesCount)
	if len(files) > 0 {
		for i, file := range strings.Split(message.Files, ",") {
			files[i] = file
		}
	}
	return &dto.MessageResponse{
		MessageId:     messageId,
		SenderId:      senderId,
		DestinationId: destinationId,
		IsDirect:      message.IsDirect,
		Body:          message.Body,
		IsEdited:      message.IsEdited,
		IsDeleted:     message.IsDeleted,
		CreatedAt:     message.CreatedAt,
		Files:         files,
	}
}

type ConnectionInfo struct {
	ChatId       uuid.UUID
	WsConnection *websocket.Conn
}
