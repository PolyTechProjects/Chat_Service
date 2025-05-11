package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"example.com/messaging/src/internal/client"
	"example.com/messaging/src/internal/dto"
	"example.com/messaging/src/internal/models"
	"example.com/messaging/src/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageHistoryService struct {
	messageRepository *repository.MessageRepository
	chatClient        *client.ChatGRPCClient
}

func NewMessageHistoryService(messageRepository *repository.MessageRepository, chatClient *client.ChatGRPCClient) *MessageHistoryService {
	return &MessageHistoryService{
		messageRepository: messageRepository,
		chatClient:        chatClient,
	}
}

func (m *MessageHistoryService) GetDirectHistory(destinationId uuid.UUID, userId uuid.UUID) ([]models.Message, error) {
	messages, err := m.messageRepository.GetDirectMessages(userId, destinationId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (m *MessageHistoryService) GetHistory(destinationId uuid.UUID, userId uuid.UUID) ([]models.Message, error) {
	verifyPersistanceResp, err := m.chatClient.PerformVerifyUserPersistance(destinationId.String(), userId.String())
	if err != nil {
		return nil, err
	}
	if !verifyPersistanceResp.IsVerified {
		return nil, fmt.Errorf("permission denied")
	}
	messages, err := m.messageRepository.GetMessagesByDestinationId(destinationId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

type MessageService struct {
	messageRepository   *repository.MessageRepository
	chatClient          *client.ChatGRPCClient
	redisClient         *client.RedisClient
	userIdXWsConnection map[uuid.UUID]*websocket.Conn
	broadcastChannel    chan *models.Message
}

func NewMessageService(messageRepository *repository.MessageRepository, chatClient *client.ChatGRPCClient, redisClient *client.RedisClient) *MessageService {
	return &MessageService{
		messageRepository:   messageRepository,
		chatClient:          chatClient,
		redisClient:         redisClient,
		userIdXWsConnection: make(map[uuid.UUID]*websocket.Conn),
		broadcastChannel:    make(chan *models.Message),
	}
}

func (s *MessageService) ReadMessages(wsConnection *websocket.Conn, userId uuid.UUID, destinationId uuid.UUID) error {
	var cerr error
	cerr = s.redisClient.ConnectUser(userId)
	if cerr != nil {
		return cerr
	}
	s.userIdXWsConnection[userId] = wsConnection
	for {
		_, payload, err := wsConnection.ReadMessage()
		if err != nil {
			cerr = err
			break
		}

		req := &dto.MessageRequest{}
		err = json.Unmarshal(payload, req)
		if err != nil {
			cerr = err
			break
		}

		if len(req.Files) > 0 {
			verifyUserActionResp, err := s.chatClient.PerformVerifyUserAction(destinationId.String(), userId.String(), "SEND_FILE")
			if err != nil {
				cerr = err
				break
			}
			if !verifyUserActionResp.IsVerified {
				cerr = fmt.Errorf("permission denied")
				break
			}
		}

		message, err := models.MapRequestToMessage(req, userId, destinationId)
		if err != nil {
			cerr = err
			break
		}
		err = s.messageRepository.SaveMessage(message)
		if err != nil {
			cerr = err
			break
		}

		s.broadcastChannel <- message
		s.redisClient.SendToMessageChannel(message)
	}
	delete(s.userIdXWsConnection, userId)
	err := s.redisClient.DisconnectUser(userId)
	if err != nil {
		cerr = err
	}
	return cerr
}

func (s *MessageService) BroadcastMessageGoChannel() {
	for message := range s.broadcastChannel {
		s.doBroadcastMessage(message)
	}
}

func (s *MessageService) BroadcastMessageRedisChannel(message *models.Message) {
	s.doBroadcastMessage(message)
}

func (s *MessageService) doBroadcastMessage(message *models.Message) {
	getChatResponse, err := s.chatClient.PerformGetChat(message.DestinationId.String(), message.SenderId.String())
	if err != nil {
		slog.Error(err.Error())
		return
	}
	notificationEvent := &dto.NewMessageNotificationEvent{
		MessageId:  message.Id.String(),
		SenderId:   message.SenderId.String(),
		Body:       message.Body,
		FilesCount: len(message.Files),
	}
	for _, participant := range getChatResponse.ParticipantsIds {
		participantId := uuid.MustParse(participant)
		wsConnection, ok := s.userIdXWsConnection[participantId]
		if ok {
			messageResp := models.MapMessageToResponse(message)
			messageJson, err := json.Marshal(messageResp)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			err = wsConnection.WriteJSON(messageJson)
			if err != nil {
				slog.Error(fmt.Sprintf("Disconnect user %s due to error %v", participantId, err.Error()))
				wsConnection.Close()
				delete(s.userIdXWsConnection, participantId)
			}
		} else {
			isConnected, err := s.redisClient.IsUserConnected(participantId)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			if !isConnected {
				s.redisClient.SendToNotificationChannel(notificationEvent)
			}
		}
	}
}
