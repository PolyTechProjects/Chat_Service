package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/user_mgmt/src/config"
	"example.com/user_mgmt/src/gen/go/auth"
	"example.com/user_mgmt/src/internal/dto"
	"example.com/user_mgmt/src/internal/service"
	"github.com/go-redis/redis"
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

func (authClient *AuthGRPCClient) PerformRefresh(r *http.Request) (*auth.RefreshResponse, error) {
	refreshCookie, err := r.Cookie("X-Refresh-Token")
	if err != nil {
		slog.Error("PerformRefresh failed: " + err.Error())
		return nil, fmt.Errorf("PerformRefresh failed: " + err.Error())
	}
	refreshToken := refreshCookie.Value
	ctx := r.Context()
	refreshRequest := &auth.RefreshRequest{RefreshToken: refreshToken}
	return authClient.Refresh(ctx, refreshRequest)
}

type RedisClient struct {
	Client                   *redis.Client
	createAccountChannelName string
	deleteAccountChannelName string
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.InnerPort),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	return &RedisClient{
		Client:                   client,
		createAccountChannelName: cfg.Redis.CreateAccountChannelName,
		deleteAccountChannelName: cfg.Redis.DeleteAccountChannelName,
	}
}

func (c *RedisClient) SubscribeToCreateAccountChannel(service *service.UserMgmtService) {
	pubsub := c.Client.Subscribe(c.createAccountChannelName)
	defer pubsub.Close()
	accountCreatedEvent := &dto.AccountCreatedEvent{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), accountCreatedEvent)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		service.CreateUser(accountCreatedEvent)
	}
}

func (c *RedisClient) SubscribeToDeleteAccountChannel(service *service.UserMgmtService) {
	pubsub := c.Client.Subscribe(c.deleteAccountChannelName)
	defer pubsub.Close()
	accountDeletedEvent := &dto.AccountDeletedEvent{}
	for msg := range pubsub.Channel() {
		slog.Debug("New msg received: " + msg.Payload)
		err := json.Unmarshal([]byte(msg.Payload), accountDeletedEvent)
		if err != nil {
			slog.Error("Error while unmarshalling: " + err.Error())
		}
		service.DeleteUser(accountDeletedEvent)
	}
}
