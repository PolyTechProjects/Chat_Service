package client

import (
	"context"
	"fmt"

	"example.com/main/src/config"
	"example.com/main/src/gen/go/user_mgmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserMgmtGRPCClient struct {
	client user_mgmt.UserMgmtClient
}

func New(cfg *config.Config) *UserMgmtGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.UserMgmt.UserMgmtHost, cfg.UserMgmt.UserMgmtPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return &UserMgmtGRPCClient{client: user_mgmt.NewUserMgmtClient(conn)}
}

func (c *UserMgmtGRPCClient) PerformAddUser(ctx context.Context, userId string, name string) (*user_mgmt.UserResponse, error) {
	return c.client.AddUser(ctx, &user_mgmt.AddUserRequest{
		UserId: userId,
		Name:   name,
	})
}
