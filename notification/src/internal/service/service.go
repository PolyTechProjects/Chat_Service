package service

import (
	"encoding/json"
	"errors"
	"log/slog"

	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/dto"
	"example.com/notification/src/internal/mail"
	"example.com/notification/src/internal/repository"
	"example.com/notification/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var mailsXMessageType = map[string]mail.MailBuilder{
	"direct_new_message": mail.NewDirectNewMessageMailBuilder(),
	"chat_new_message":   mail.NewChatNewMessageMailBuiler(),
	"default":            mail.NewDefaultMailBuilder(),
}

type NotificationService struct {
	notificationRepository *repository.NotificationRepository
	subscriptionRepository *repository.SubscriptionRepository
	authClient             *client.AuthGRPCClient
	usersClient            *client.UsersGRPCClient
	chatClient             *client.ChatGRPCClient
	redisClient            *client.RedisClient
	smtpClient             *client.SmtpClient
}

func NewNotificationService(
	notificationRepository *repository.NotificationRepository,
	subscriptionRepository *repository.SubscriptionRepository,
	authClient *client.AuthGRPCClient,
	usersClient *client.UsersGRPCClient,
	chatClient *client.ChatGRPCClient,
	redisClient *client.RedisClient,
	smtpClient *client.SmtpClient,
) *NotificationService {
	return &NotificationService{
		notificationRepository: notificationRepository,
		subscriptionRepository: subscriptionRepository,
		authClient:             authClient,
		usersClient:            usersClient,
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
	if err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
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
	} else if err != nil {
		slog.Error("Failed to find subscription: " + err.Error())
		return err
	}

	var messageType string
	var mailSubjectOpts *mail.MailSubjectOpts
	var mailBodyOpts *mail.MailBodyOpts
	if event.IsDirect {
		messageType = "direct_new_message"
		userResp, err := s.usersClient.PerformGetUser(senderId.String())
		if err != nil {
			return err
		}
		mailSubjectOpts = &mail.MailSubjectOpts{
			SenderName: userResp.Name,
		}
		mailBodyOpts = &mail.MailBodyOpts{
			SenderName: userResp.Name,
			FilesCount: event.FilesCount,
			Body:       event.Body,
		}
	} else {
		messageType = "chat_new_message"
		chatResp, err := s.chatClient.PerformGetChatAndUserNames(chatId.String(), senderId.String())
		if err != nil {
			return err
		}
		mailSubjectOpts = &mail.MailSubjectOpts{
			ChatName:   chatResp.ChatName,
			SenderName: chatResp.UserName,
		}
		mailBodyOpts = &mail.MailBodyOpts{
			SenderName: chatResp.UserName,
			FilesCount: event.FilesCount,
			Body:       event.Body,
		}
	}
	loginResp, err := s.authClient.PerformGetLogin(event.ReceiverId)
	if err != nil {
		return err
	}
	mailSubject := mailsXMessageType[messageType].BuildSubject(*&mailSubjectOpts)
	mailBody := mailsXMessageType[messageType].BuildBody(*&mailBodyOpts)
	readyMail := mailsXMessageType[messageType].Build(loginResp.Login, mailSubject, mailBody)
	err = s.smtpClient.SendEmail(readyMail)
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

func (s *NotificationService) SendEmail(event *dto.SendEmailEvent) error {
	mail := mailsXMessageType["default"].Build(event.To, event.Subject, event.Body)
	err := s.smtpClient.SendEmail(mail)
	if err != nil {
		slog.Error("Failed to send email: " + err.Error())
		return err
	}
	return nil
}
