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

func (s *MessageHistoryService) GetDirectHistory(destinationId uuid.UUID, userId uuid.UUID) ([]models.Message, error) {
	messages, err := s.messageRepository.GetDirectMessages(userId, destinationId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageHistoryService) GetHistory(destinationId uuid.UUID, userId uuid.UUID) ([]models.Message, error) {
	verifyPersistanceResp, err := s.chatClient.PerformVerifyUserPersistance(destinationId.String(), userId.String())
	if err != nil {
		return nil, err
	}
	if !verifyPersistanceResp.IsVerified {
		return nil, fmt.Errorf("permission denied")
	}
	messages, err := s.messageRepository.GetMessagesByDestinationId(destinationId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageHistoryService) DeleteMessage(messageId uuid.UUID, userId uuid.UUID) error {
	message, err := s.messageRepository.FindById(messageId)
	if err != nil {
		return err
	}
	if message.IsDirect {
		if message.SenderId != userId {
			return fmt.Errorf("permission denied")
		}
		return s.doDeleteMessage(message)
	}
	verifyPersistanceResp, err := s.chatClient.PerformVerifyUserPersistance(message.DestinationId.String(), userId.String())
	if err != nil {
		return err
	}
	if !verifyPersistanceResp.IsVerified {
		return fmt.Errorf("permission denied")
	}
	if message.SenderId == userId {
		verifyAction, err := s.chatClient.PerformVerifyUserAction(message.DestinationId.String(), userId.String(), "CAN_DELETE_MESSAGE")
		if err != nil {
			return err
		}
		if !verifyAction.IsVerified {
			return fmt.Errorf("permission denied")
		}
	} else {
		verifyAction, err := s.chatClient.PerformVerifyUserActionOnSomebody(
			message.DestinationId.String(),
			userId.String(),
			message.SenderId.String(),
			"CAN_DELETE_OTHERS_MESSAGE",
		)
		if err != nil {
			return err
		}
		if !verifyAction.IsVerified {
			return fmt.Errorf("permission denied")
		}
	}
	return s.doDeleteMessage(message)
}

func (s *MessageHistoryService) EditMessage(messageId uuid.UUID, newBody string, userId uuid.UUID) error {
	message, err := s.messageRepository.FindById(messageId)
	if err != nil {
		return err
	}
	if message.IsDirect {
		if message.SenderId != userId {
			return fmt.Errorf("permission denied")
		}
		return s.doEditMessage(message, newBody)
	}
	verifyPersistanceResp, err := s.chatClient.PerformVerifyUserPersistance(message.DestinationId.String(), userId.String())
	if err != nil {
		return err
	}
	if !verifyPersistanceResp.IsVerified {
		return fmt.Errorf("permission denied")
	}
	if message.SenderId == userId {
		verifyAction, err := s.chatClient.PerformVerifyUserAction(message.DestinationId.String(), userId.String(), "CAN_EDIT_MESSAGE")
		if err != nil {
			return err
		}
		if !verifyAction.IsVerified {
			return fmt.Errorf("permission denied")
		}
	} else {
		verifyAction, err := s.chatClient.PerformVerifyUserActionOnSomebody(
			message.DestinationId.String(),
			userId.String(),
			message.SenderId.String(),
			"CAN_EDIT_OTHERS_MESSAGE",
		)
		if err != nil {
			return err
		}
		if !verifyAction.IsVerified {
			return fmt.Errorf("permission denied")
		}
	}
	return s.doEditMessage(message, newBody)
}

func (s *MessageHistoryService) doDeleteMessage(message *models.Message) error {
	message.IsDeleted = true
	err := s.messageRepository.SaveMessage(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *MessageHistoryService) doEditMessage(message *models.Message, newBody string) error {
	message.Body = newBody
	err := s.messageRepository.SaveMessage(message)
	if err != nil {
		return err
	}
	return nil
}

type MessageService struct {
	messageRepository  *repository.MessageRepository
	chatClient         *client.ChatGRPCClient
	redisClient        *client.RedisClient
	connectionRegistry map[uuid.UUID]*models.ConnectionInfo
}

func NewMessageService(messageRepository *repository.MessageRepository, chatClient *client.ChatGRPCClient, redisClient *client.RedisClient) *MessageService {
	return &MessageService{
		messageRepository:  messageRepository,
		chatClient:         chatClient,
		redisClient:        redisClient,
		connectionRegistry: make(map[uuid.UUID]*models.ConnectionInfo),
	}
}

func (s *MessageService) ReadMessages(wsConnection *websocket.Conn, userId uuid.UUID, destinationId uuid.UUID) error {
	err := s.redisClient.ConnectUser(userId)
	if err != nil {
		return err
	}
	s.connectionRegistry[userId] = &models.ConnectionInfo{
		ChatId:       destinationId,
		WsConnection: wsConnection,
	}
	stop := make(chan struct{})
	for {
		select {
		case <-stop:
			slog.Info("Closing connection")
			s.closeConnection(userId)
		default:
			_, payload, err := wsConnection.ReadMessage()
			if err != nil {
				slog.Error("MessageServiceReadMessages failed: " + err.Error())
				s.closeConnection(userId)
				return err
			}
			verifyAction, err := s.chatClient.PerformVerifyUserAction(destinationId.String(), userId.String(), "CAN_WRITE_MESSAGE")
			if err != nil {
				slog.Error("MessageServiceReadMessages failed: " + err.Error())
				s.closeConnection(userId)
				return err
			}
			if !verifyAction.IsVerified {
				slog.Error("MessageServiceReadMessages failed: " + fmt.Errorf("permission denied").Error())
				s.closeConnection(userId)
				return err
			}
			req := &dto.MessageRequest{}
			err = json.Unmarshal(payload, req)
			if err != nil {
				slog.Error("MessageServiceReadMessages failed: " + err.Error())
				s.closeConnection(userId)
				return err
			}
			if len(req.Files) > 0 && !req.IsDirect {
				verifyAction, err := s.chatClient.PerformVerifyUserAction(destinationId.String(), userId.String(), "CAN_SEND_FILE")
				if err != nil {
					slog.Error("MessageServiceReadMessages failed: " + err.Error())
					s.closeConnection(userId)
					return err
				}
				if !verifyAction.IsVerified {
					slog.Error("MessageServiceReadMessages failed: " + fmt.Errorf("permission denied").Error())
					s.closeConnection(userId)
					return err
				}
			}
			message, err := models.MapRequestToMessage(req, userId, destinationId)
			if err != nil {
				slog.Error("MessageServiceReadMessages failed: " + err.Error())
				s.closeConnection(userId)
				return err
			}
			err = s.messageRepository.SaveMessage(message)
			if err != nil {
				slog.Error("MessageServiceReadMessages failed: " + err.Error())
				s.closeConnection(userId)
				return err
			}
			s.redisClient.SendToMessageChannel(message)
		}
	}
}

func (s *MessageService) BroadcastMessageRedisChannel(message *models.Message) {
	s.doBroadcastMessage(message)
}

func (s *MessageService) doBroadcastMessage(message *models.Message) {
	notificationEvent := &dto.NewMessageNotificationEvent{
		MessageId:     message.Id.String(),
		SenderId:      message.SenderId.String(),
		DestinationId: message.DestinationId.String(),
		Body:          message.Body,
		IsDirect:      message.IsDirect,
		FilesCount:    len(message.Files),
	}
	if message.IsDirect {
		s.broadcast(message, message.DestinationId, notificationEvent)
		return
	}
	getChatResponse, err := s.chatClient.PerformGetChat(message.DestinationId.String(), message.SenderId.String())
	if err != nil {
		slog.Error("MessageService doBroadcastMessage failed: " + err.Error())
		s.closeConnection(message.SenderId)
		return
	}
	for _, participant := range getChatResponse.ParticipantsIds {
		participantId := uuid.MustParse(participant)
		if participantId == message.SenderId {
			continue
		}
		s.broadcast(message, participantId, notificationEvent)
	}
}

func (s *MessageService) broadcast(message *models.Message, receiverId uuid.UUID, notificationEvent *dto.NewMessageNotificationEvent) {
	notificationEvent.ReceiverId = receiverId.String()
	notificationEvent.EventId = uuid.NewString()
	connectionInfo, ok := s.connectionRegistry[receiverId]
	// if opened on this instance and this endpoint - write to connection
	// if opened on this instance and other endpoint - send notification
	// if opened on other instance and this endpoint - do nothing
	// if opened on other instance and other endpoint - send notification
	// if not opened at all - send notification
	if ok {
		if connectionInfo.ChatId == message.DestinationId {
			messageResp := models.MapMessageToResponse(message)
			err := connectionInfo.WsConnection.WriteJSON(messageResp)
			if err != nil {
				slog.Error("MessageService broadcast failed: " + err.Error())
				s.closeConnection(receiverId)
			}
			return
		} else {
			s.redisClient.SendToNotificationChannel(notificationEvent)
			return
		}
	} else {
		isConnected, err := s.redisClient.IsUserConnected(receiverId)
		if err != nil {
			slog.Error("MessageService broadcast failed: " + err.Error())
			s.closeConnection(receiverId)
		}
		if !isConnected {
			s.redisClient.SendToNotificationChannel(notificationEvent)
		}
	}
}

func (s *MessageService) closeConnection(userId uuid.UUID) {
	slog.Debug("Closing connection...")
	s.connectionRegistry[userId].WsConnection.Close()
	delete(s.connectionRegistry, userId)
	s.redisClient.DisconnectUser(userId)
}
