package models

import (
	"example.com/chat-app/src/internal/dto"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type ChatRoomXUser struct {
	gorm.Model
	ChatRoomId uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	UserId     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
}

type MessageIdXFileId struct {
	MessageId uuid.UUID `json:"messageId"`
	FileId    uuid.UUID `json:"fileId"`
}

type Message struct {
	Id         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	SenderId   uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	ChatRoomId uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Body       string
	CreatedAt  uint64
	WithMedia  int
	Metadata   dto.Metadata `gorm:"serializer:json"`
}

func MapRequestToMessage(req *dto.MessageRequest) (*Message, error) {
	messageUUID, err := uuid.Parse(req.MessageId)
	if err != nil {
		return nil, err
	}
	senderUUID, err := uuid.Parse(req.SenderId)
	if err != nil {
		return nil, err
	}
	chatRoomUUID, err := uuid.Parse(req.ChatRoomId)
	if err != nil {
		return nil, err
	}
	return &Message{
		Id:         messageUUID,
		SenderId:   senderUUID,
		ChatRoomId: chatRoomUUID,
		Body:       req.Body,
		CreatedAt:  req.CreatedAt,
		WithMedia:  req.WithMedia,
	}, nil
}

func MapMessageToResponse(message *Message) *dto.MessageResponse {
	senderId := message.SenderId.String()
	chatRoomId := message.ChatRoomId.String()
	return &dto.MessageResponse{
		SenderId:   senderId,
		ChatRoomId: chatRoomId,
		Body:       message.Body,
		CreatedAt:  message.CreatedAt,
		Metadata:   message.Metadata,
	}
}
