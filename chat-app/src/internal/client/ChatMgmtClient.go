package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/config"
	chatMgmt "example.com/chat-app/src/gen/go/chat_mgmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

func (chatMgmtClient *ChatMgmtGRPCClient) PerformGetChatUsers(chatID, accessToken string, refreshToken string) ([]uuid.UUID, error) {
	md := metadata.Pairs("authorization", accessToken)
	md.Append("x-refresh-token", refreshToken)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := chatMgmtClient.GetChat(ctx, &chatMgmt.GetChatRequest{ChatId: chatID})
	if err != nil {
		return nil, err
	}
	var userIds []uuid.UUID
	for _, id := range resp.ParticipantsIds {
		userId, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}
	for _, id := range resp.AdminsIds {
		userId, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, userId)
	}
	return userIds, nil
}
