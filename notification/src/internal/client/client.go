package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/notification/src/config"
	"example.com/notification/src/gen/go/auth"
	"example.com/notification/src/gen/go/chat"
	"example.com/notification/src/gen/go/users"
	"example.com/notification/src/internal/dto"
	"example.com/notification/src/internal/mail"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gopkg.in/gomail.v2"
)

type AuthGRPCClient struct {
	auth.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth: " + connectionUrl)
	return &AuthGRPCClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(r *http.Request) (*auth.AuthorizeResponse, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		slog.Error("PerformAuthorize failed: No access token provided")
		return nil, fmt.Errorf("PerformAuthorize failed: No access token provided")
	}
	ctx := metadata.AppendToOutgoingContext(r.Context(), "Authorization", "Bearer "+accessToken)
	return authClient.Authorize(ctx, &auth.AuthorizeRequest{})
}

func (authClient *AuthGRPCClient) PerformExtractUserId(r *http.Request) (*auth.ExtractUserIdResponse, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		slog.Error("PerformExtractUserId failed: No access token provided")
		return nil, fmt.Errorf("PerformExtractUserId failed: No access token provided")
	}
	ctx := metadata.AppendToOutgoingContext(r.Context(), "Authorization", "Bearer "+accessToken)
	return authClient.ExtractUserId(ctx, &auth.ExtractUserIdRequest{})
}

func (c *AuthGRPCClient) PerformGetLogin(userId string) (*auth.GetLoginResponse, error) {
	getLoginResponse, err := c.AuthClient.GetLogin(context.Background(), &auth.GetLoginRequest{UserId: userId})
	if err != nil {
		slog.Error("PerformGetLogin failed : " + err.Error())
		return nil, err
	}
	return getLoginResponse, nil
}

type UsersGRPCClient struct {
	users.UsersClient
}

func NewUsersClient(cfg *config.Config) *UsersGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Users.UsersHost, cfg.Users.UsersPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Users: " + connectionUrl)
	return &UsersGRPCClient{users.NewUsersClient(conn)}
}

func (c *UsersGRPCClient) PerformGetUser(userId string) (*users.UserResponse, error) {
	getUserResponse, err := c.UsersClient.GetUser(context.Background(), &users.GetUserRequest{UserId: userId})
	if err != nil {
		slog.Error("PerformGetUser failed : " + err.Error())
		return nil, err
	}
	return getUserResponse, nil
}

type ChatGRPCClient struct {
	chat.ChatClient
}

func NewChatClient(cfg *config.Config) *ChatGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Chat.ChatHost, cfg.Chat.ChatPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Chat: " + connectionUrl)
	return &ChatGRPCClient{chat.NewChatClient(conn)}
}

func (c *ChatGRPCClient) PerformGetChatAndUserNames(chatId string, userId string) (*chat.GetChatAndUserNamesResponse, error) {
	getChatAndUserNamesResponse, err := c.ChatClient.GetChatAndUserNames(context.Background(), &chat.GetChatAndUserNamesRequest{ChatId: chatId, UserId: userId})
	if err != nil {
		slog.Error("PerformGetChatAndUserNames failed : " + err.Error())
		return nil, err
	}
	return getChatAndUserNamesResponse, nil
}

type RedisClient struct {
	Client                  *redis.Client
	NotificationChannelName string
	SubscriptionChannelName string
	SendEmailChannelName    string
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.InnerPort),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	return &RedisClient{
		Client:                  client,
		NotificationChannelName: cfg.Redis.NotificationChannelName,
		SubscriptionChannelName: cfg.Redis.SubscriptionChannelName,
		SendEmailChannelName:    cfg.Redis.SendEmailChannelName,
	}
}

func (r *RedisClient) SubscribeToNotificationChannel(handler func(event *dto.NewMessageNotificationEvent) error) {
	pubsub := r.Client.Subscribe(r.NotificationChannelName)
	defer pubsub.Close()
	event := &dto.NewMessageNotificationEvent{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), event)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		err = handler(event)
		if err != nil {
			slog.Error("Error while handling event: " + err.Error())
		}
	}
}

func (r *RedisClient) SubscribeToSubscriptionChannel(handler func(event *dto.NewSubscriptionNotificationEvent) error) {
	pubsub := r.Client.Subscribe(r.SubscriptionChannelName)
	defer pubsub.Close()
	event := &dto.NewSubscriptionNotificationEvent{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), event)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		err = handler(event)
		if err != nil {
			slog.Error("Error while handling event: " + err.Error())
		}
	}
}

func (r *RedisClient) SubscribeToSendEmailChannel(handler func(event *dto.SendEmailEvent) error) {
	pubsub := r.Client.Subscribe(r.SendEmailChannelName)
	defer pubsub.Close()
	message := &dto.SendEmailEvent{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), message)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		err = handler(message)
		if err != nil {
			slog.Error("Error while handling event: " + err.Error())
		}
	}
}

type SmtpClient struct {
	User   string
	Dialer *gomail.Dialer
}

func NewSmtpClient(cfg *config.Config) *SmtpClient {
	return &SmtpClient{
		User:   cfg.Smtp.User,
		Dialer: gomail.NewDialer(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.User, cfg.Smtp.Password),
	}
}

func (s *SmtpClient) SendEmail(mail *mail.Mail) error {
	message := gomail.NewMessage()
	message.SetHeader("From", s.User)
	message.SetHeader("To", mail.To)
	message.SetHeader("Subject", mail.Subject)
	message.SetBody("text/plain", mail.Body)
	s.Dialer.DialAndSend(message)
	return nil
}
