package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/internal/dto"
	"example.com/chat-app/src/internal/models"
	"example.com/chat-app/src/internal/repository"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageHistoryService struct {
	messageRepository *repository.MessageRepository
}

func NewMessageHistoryService(messageRepository *repository.MessageRepository) *MessageHistoryService {
	return &MessageHistoryService{
		messageRepository: messageRepository,
	}
}

func (m *MessageHistoryService) GetHistory(chatRoomId uuid.UUID) ([]models.Message, error) {
	messages, err := m.messageRepository.GetMessageByChatRoomId(chatRoomId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

type MessageService struct {
	messageRepository   *repository.MessageRepository
	fileLoadedChannel   chan *models.MessageIdXFileId
	userIdXWsConnection map[uuid.UUID]*websocket.Conn
}

func NewMessageService(messageRepository *repository.MessageRepository) *MessageService {
	return &MessageService{
		messageRepository:   messageRepository,
		fileLoadedChannel:   make(chan *models.MessageIdXFileId),
		userIdXWsConnection: make(map[uuid.UUID]*websocket.Conn),
	}
}

func (m *MessageService) ListenFileChannel() {
	var subscriber = m.messageRepository.SubscribeToRedisChannel("file-loaded-channel")
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Not Available file-loaded-channel")
	}
	slog.Info("Available file-loaded-channel")
	for {
		channel := subscriber.Channel()
		message := <-channel
		slog.Info(message.Payload)
		mf := &models.MessageIdXFileId{}
		err = json.Unmarshal([]byte(message.Payload), mf)
		if err != nil {
			slog.Error(err.Error())
		}
		m.fileLoadedChannel <- mf
	}
}

func (m *MessageService) Broadcast(userIds []uuid.UUID, message *models.Message) {
	for _, userId := range userIds {
		slog.Debug(fmt.Sprintf("Checking if user %s is connected", userId))
		wsConnection, ok := m.userIdXWsConnection[userId]
		if ok {
			slog.Debug(fmt.Sprintf("User %s is connected", userId))
			messageResp := models.MapMessageToResponse(message)
			slog.Debug(fmt.Sprintf("Sending message to %s", userId))
			err := wsConnection.WriteJSON(messageResp)
			if err != nil {
				slog.Error(fmt.Sprintf("Disconnect user %s due to error %v", userId, err.Error()))
				wsConnection.Close()
				delete(m.userIdXWsConnection, userId)
			}
		} else {
			slog.Debug(fmt.Sprintf("User %s is not connected", userId))
			slog.Debug(fmt.Sprintf("Trying to notify user %v", userId))
			m.messageRepository.PublishToRedisChannel("notification", message)
		}
	}
}

func (m *MessageService) ReadMessages(userId uuid.UUID, wsConnection *websocket.Conn, broadcastChannel chan *models.Message, token string) {
	m.userIdXWsConnection[userId] = wsConnection
	slog.Debug(fmt.Sprintf("Added wsConnection to %v", userId))

	for {
		_, payload, err := wsConnection.ReadMessage()
		if err != nil {
			slog.Error("Error has occured while reading message", err)
			break
		}

		messageReq := dto.MessageRequest{}

		err = json.Unmarshal(payload, &messageReq)
		if err != nil {
			slog.Error("Error has occured while unmarshalling message", err)
			break
		}

		message, err := models.MapRequestToMessage(&messageReq)
		if err != nil {
			slog.Error("Error has occured while mapping request to message", err)
			break
		}

		mediaReceived := 0
		if messageReq.WithMedia > 0 {
			slog.Debug("Getting files")
			for messageReq.WithMedia != mediaReceived {
				mf := <-m.fileLoadedChannel
				if mf.MessageId == message.Id {
					metadata := dto.Metadata{}
					metadata.FilePath = mf.FileId.String()
					message.Metadata = metadata
					mediaReceived++
				} else {
					m.fileLoadedChannel <- mf
				}
			}
			slog.Debug("Files received")
		}

		slog.Debug("Saving message")
		err = m.messageRepository.SaveUserMessage(message)
		if err != nil {
			slog.Error(fmt.Sprintf("Error has occured while saving message: %v", err.Error()))
			break
		}
		slog.Debug(fmt.Sprintf("Message Saved %v, %v", message.Id, message.Metadata.FilePath))

		slog.Debug("Publishing Message")
		messageWithToken := models.MessageWithToken{
			Message: *message,
			Token:   token,
		}
		bytes, err := json.Marshal(messageWithToken)
		if err != nil {
			slog.Error(fmt.Sprintf("Error has occured while marshalling message: %v", err.Error()))
			break
		}
		err = m.messageRepository.PublishToRedisChannel("message-channel", bytes)
		if err != nil {
			slog.Error(fmt.Sprintf("Error has occured while publishing message: %v", err.Error()))
			break
		}
		slog.Debug(fmt.Sprintf("Message Published %v", bytes))
		//broadcastChannel <- message
	}

	slog.Debug(fmt.Sprintf("Removing wsConnection from %v", userId))
	delete(m.userIdXWsConnection, userId)
}

func (m *MessageService) SubscribeToMessageChannel() *redis.PubSub {
	return m.messageRepository.SubscribeToRedisChannel("message-channel")
}
