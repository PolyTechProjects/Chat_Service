package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/config"
	chatMgmt "example.com/chat-app/src/gen/go/chat-mgmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatMgmtGRPCClient struct {
	chatMgmt.ChatManagementClient
}

func NewChatMgmtClient(cfg *config.Config) *ChatMgmtGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.ChatMgmt.ChatMgmtHost, cfg.ChatMgmt.ChatMgmtPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to ChatManagement")
	slog.Info(connectionUrl)
	return &ChatMgmtGRPCClient{chatMgmt.NewChatManagementClient(conn)}
}

func (chatMgmtClient *ChatMgmtGRPCClient) PerformGetChatUsers(chatID string) ([]uuid.UUID, error) {
	resp, err := chatMgmtClient.GetChatUsers(context.Background(), &chatMgmt.GetChatUsersRequest{ChatId: chatID})
	if err != nil {
		return nil, err
	}
	var userIDs []uuid.UUID
	for _, id := range resp.UserIds {
		userUUID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userUUID)
	}
	return userIDs, nil
}
