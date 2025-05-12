package service

import (
	"encoding/json"
	"errors"
	"log/slog"

	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/dto"
	"example.com/notification/src/internal/repository"
	"example.com/notification/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type NotificationService struct {
	notificationRepository *repository.NotificationRepository
	subscriptionRepository *repository.SubscriptionRepository
	authClient             *client.AuthGRPCClient
	chatClient             *client.ChatGRPCClient
	redisClient            *client.RedisClient
	smtpClient             *client.SmtpClient
}

func NewNotificationService(
	notificationRepository *repository.NotificationRepository,
	subscriptionRepository *repository.SubscriptionRepository,
	authClient *client.AuthGRPCClient,
	chatClient *client.ChatGRPCClient,
	redisClient *client.RedisClient,
	smtpClient *client.SmtpClient,
) *NotificationService {
	return &NotificationService{
		notificationRepository: notificationRepository,
		subscriptionRepository: subscriptionRepository,
		authClient:             authClient,
		chatClient:             chatClient,
		redisClient:            redisClient,
		smtpClient:             smtpClient,
	}
}

func (s *NotificationService) Unsubscribe(chatId uuid.UUID, userId uuid.UUID) error {
	return s.subscriptionRepository.DeleteSubscriptionsByChatIdAndUserId(chatId, userId)
}

func (s *NotificationService) Subscribe(event *dto.NewSubscriptionNotificationEvent) error {
	return s.subscriptionRepository.AddSubscription(&models.Subscription{
		ChatId: event.ChatId,
		UserId: event.UserId,
	})
}

func (s *NotificationService) SendNotification(event *dto.NewMessageNotificationEvent) error {
	_, err := s.notificationRepository.FindById(uuid.MustParse(event.EventId))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		slog.Error("Failed to find notification: " + err.Error())
		return err
	}

	chatId, err := uuid.Parse(event.DestinationId)
	if err != nil {
		slog.Error("Failed to parse chat id: " + err.Error())
		return err
	}
	senderId, err := uuid.Parse(event.SenderId)
	if err != nil {
		slog.Error("Failed to parse sender id: " + err.Error())
		return err
	}
	_, err = s.subscriptionRepository.FindSubscriptionsByChatIdAndUserId(chatId, senderId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		slog.Error("Failed to find subscription: " + err.Error())
		return err
	}

	chatResp, err := s.chatClient.PerformGetChatAndUserNames(chatId.String(), senderId.String())
	if err != nil {
		return err
	}
	loginResp, err := s.authClient.PerformGetLogin(event.ReceiverId)
	if err != nil {
		return err
	}
	err = s.smtpClient.SendEmail(chatResp.UserName, loginResp.Login, chatResp.ChatName, event.Body, event.FilesCount, event.IsDirect)
	if err != nil {
		slog.Error("Failed to send email: " + err.Error())
		return err
	}

	e, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal event: " + err.Error())
		return err
	}
	eventJson := json.RawMessage(e)
	notification := &models.Notification{
		Id:         uuid.MustParse(event.EventId),
		ReceiverId: uuid.MustParse(event.ReceiverId),
		Type:       models.MAIL,
		Body:       &eventJson,
	}
	s.notificationRepository.AddNotification(notification)
	return nil
}
