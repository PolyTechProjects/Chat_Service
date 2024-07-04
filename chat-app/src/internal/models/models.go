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
	Metadata   dto.Metadata `gorm:"type:jsonb"`
}

type MessageWithTokens struct {
	Message      Message `json:"message"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
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
	messageId := message.Id.String()
	senderId := message.SenderId.String()
	chatRoomId := message.ChatRoomId.String()
	return &dto.MessageResponse{
		MessageId:  messageId,
		SenderId:   senderId,
		ChatRoomId: chatRoomId,
		Body:       message.Body,
		CreatedAt:  message.CreatedAt,
		Metadata:   message.Metadata,
	}
}

type ErrorMessageResponse struct {
	Error string `json:"error"`
}

type ReadyMessage struct {
	Message      Message     `json:"message"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ReceiversIds []uuid.UUID `json:"receivers_ids"`
}
