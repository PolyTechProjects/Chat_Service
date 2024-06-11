package service

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/notification/src/config"
	"example.com/notification/src/internal/repository"
	"example.com/notification/src/models"
	"firebase.google.com/go/v4/messaging"
	"github.com/appleboy/go-fcm"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type NotificationService struct {
	userIdXDeviceTokenRepository *repository.UserIdXDeviceTokenRepository
	fcmClient                    *fcm.Client
}

func NewNotificationService(userIdXDeviceTokenRepository *repository.UserIdXDeviceTokenRepository, cfg *config.Config) *NotificationService {
	fcmClient, err := fcm.NewClient(context.Background(), fcm.WithCredentialsFile(cfg.Fcm.PathToPrivateKeyFile))
	if err != nil {
		slog.Error("failed to connect: " + err.Error())
	}
	return &NotificationService{
		userIdXDeviceTokenRepository: userIdXDeviceTokenRepository,
		fcmClient:                    fcmClient,
	}
}

func (ns *NotificationService) SubscribeToRedisChannel(channelName string) *redis.PubSub {
	return ns.userIdXDeviceTokenRepository.SubscribeToRedisChannel(channelName)
}

func (ns *NotificationService) NotifyUsers(receiversIds []uuid.UUID, messageTimestamp uint64, messageBody string, name string, avatar string) error {
	slog.Info("Sending notification")
	users, _ := ns.userIdXDeviceTokenRepository.GetByUserIds(receiversIds)
	deviceTokens := make([]string, 0)
	for _, user := range users {
		deviceTokens = append(deviceTokens, user.DeviceToken)
	}
	message := messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: name,
			Body:  messageBody,
		},
		Tokens: deviceTokens,
		Data: map[string]string{
			"message_timestamp": fmt.Sprintf("%v", messageTimestamp),
			"message_body":      messageBody,
			"name":              name,
			"avatar":            avatar,
		},
	}
	ns.fcmClient.SendMulticast(context.Background(), &message)
	return nil
}

func (ns *NotificationService) BindDeviceToUser(userId uuid.UUID, deviceToken string) error {
	userIdXDeviceToken := &models.UserIdXDeviceToken{UserId: userId, DeviceToken: deviceToken}
	return ns.userIdXDeviceTokenRepository.BindDeviceTokenToUser(userIdXDeviceToken)
}

func (ns *NotificationService) UnbindDeviceFromUser(userId uuid.UUID, deviceToken string) error {
	return ns.userIdXDeviceTokenRepository.UnbindDeviceTokenFromUser(userId, deviceToken)
}

func (ns *NotificationService) DeleteUser(userId uuid.UUID) error {
	return ns.userIdXDeviceTokenRepository.DeleteUser(userId)
}

func (ns *NotificationService) UpdateOldDeviceOnUser(userId uuid.UUID, oldDeviceToken string, newDeviceToken string) error {
	return ns.userIdXDeviceTokenRepository.UpdateDeviceTokensByUserId(userId, oldDeviceToken, newDeviceToken)
}
