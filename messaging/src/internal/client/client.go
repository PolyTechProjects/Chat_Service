package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/messaging/src/config"
	"example.com/messaging/src/gen/go/auth"
	"example.com/messaging/src/gen/go/chat"
	"example.com/messaging/src/internal/dto"
	"example.com/messaging/src/internal/models"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

func (c *ChatGRPCClient) PerformGetChat(chatId string, userId string) (*chat.ChatResponse, error) {
	getChatResponse, err := c.ChatClient.GetChat(context.Background(), &chat.GetChatRequest{ChatId: chatId, UserId: userId})
	if err != nil {
		slog.Error("PerformGetChat failed : " + err.Error())
		return nil, err
	}
	return getChatResponse, nil
}

func (c *ChatGRPCClient) PerformVerifyUserAction(chatId string, userId string, action string) (*chat.VerifyUserActionResponse, error) {
	verifyRequest := &chat.VerifyUserActionRequest{ChatId: chatId, UserId: userId, Action: action}
	verifyResponse, err := c.ChatClient.VerifyUserAction(context.Background(), verifyRequest)
	if err != nil {
		slog.Error("PerformVerifyUserAction failed : " + err.Error())
		return nil, err
	}
	return verifyResponse, nil
}

func (c *ChatGRPCClient) PerformVerifyUserPersistance(chatId string, userId string) (*chat.VerifyUserPersistanceResponse, error) {
	verifyRequest := &chat.VerifyUserPersistanceRequest{ChatId: chatId, UserId: userId}
	verifyResponse, err := c.ChatClient.VerifyUserPersistance(context.Background(), verifyRequest)
	if err != nil {
		slog.Error("PerformVerifyUserPersistance failed : " + err.Error())
		return nil, err
	}
	return verifyResponse, nil
}

type RedisClient struct {
	Client                  *redis.Client
	MessagesChannelName     string
	NotificationChannelName string
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.InnerPort),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	return &RedisClient{
		Client:                  client,
		MessagesChannelName:     cfg.Redis.MessagesChannelName,
		NotificationChannelName: cfg.Redis.NotificationChannelName,
	}
}

func (r *RedisClient) SubscribeToMessageChannel(handler func(message *models.Message)) {
	pubsub := r.Client.Subscribe(r.MessagesChannelName)
	defer pubsub.Close()
	message := &models.Message{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), message)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		handler(message)
	}
}

func (r *RedisClient) SendToMessageChannel(messageEntity *models.Message) error {
	message, err := json.Marshal(messageEntity)
	if err != nil {
		slog.Error("Failed to marshal message: " + err.Error())
		return err
	}
	_, err = r.Client.Publish(r.MessagesChannelName, message).Result()
	if err != nil {
		slog.Error("Failed to publish message: " + err.Error())
		return err
	}
	return nil
}

func (r *RedisClient) SendToNotificationChannel(notificationEvent *dto.NewMessageNotificationEvent) error {
	event, err := json.Marshal(notificationEvent)
	if err != nil {
		slog.Error("Failed to marshal event: " + err.Error())
		return err
	}
	_, err = r.Client.Publish(r.NotificationChannelName, event).Result()
	if err != nil {
		slog.Error("Failed to publish event: " + err.Error())
		return err
	}
	return nil
}

func (r *RedisClient) ConnectUser(userId uuid.UUID) error {
	_, err := r.Client.Set(userId.String(), "connected", 0).Result()
	if err != nil {
		slog.Error("RedisClientConnectUser failed : " + err.Error())
		return err
	}
	return nil
}

func (r *RedisClient) IsUserConnected(userId uuid.UUID) (bool, error) {
	_, err := r.Client.Exists(userId.String()).Result()
	if err != nil {
		slog.Error("RedisClientIsUserConnected failed : " + err.Error())
		return false, err
	}
	return true, nil
}

func (r *RedisClient) DisconnectUser(userId uuid.UUID) error {
	_, err := r.Client.Del(userId.String()).Result()
	if err != nil {
		slog.Error("RedisClientDisconnectUser failed : " + err.Error())
		return err
	}
	return nil
}
