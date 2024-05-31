package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/config"
	chanMgmt "example.com/chat-app/src/gen/go/channel-mgmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChanMgmtGRPCClient struct {
	chanMgmt.ChannelManagementClient
}

func NewChanMgmtClient(cfg *config.Config) *ChanMgmtGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.ChanMgmt.ChanMgmtHost, cfg.ChanMgmt.ChanMgmtPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to ChanManagement")
	slog.Info(connectionUrl)
	return &ChanMgmtGRPCClient{chanMgmt.NewChannelManagementClient(conn)}
}

func (chanMgmtClient *ChanMgmtGRPCClient) PerformGetChanUsers(channelID string) ([]uuid.UUID, error) {
	resp, err := chanMgmtClient.GetChanUsers(context.Background(), &chanMgmt.GetChanUsersRequest{ChannelId: channelID})
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
