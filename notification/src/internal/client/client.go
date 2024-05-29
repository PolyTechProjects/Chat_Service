package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/notification/src/config"
	"example.com/notification/src/gen/go/sso"
	user_mgmt "example.com/notification/src/gen/go/user-mgmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	sso.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthClient{sso.NewAuthClient(conn)}
}

func (authClient *AuthClient) PerformAuthorize(ctx context.Context, token string) (*sso.AuthorizeResponse, error) {
	return authClient.Authorize(ctx, &sso.AuthorizeRequest{Token: token})
}

type UserMgmtClient struct {
	user_mgmt.UserMgmtClient
}

func NewUserMgmtClient(cfg *config.Config) *UserMgmtClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.UserMgmt.UserMgmtHost, cfg.UserMgmt.UserMgmtPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return &UserMgmtClient{user_mgmt.NewUserMgmtClient(conn)}
}

func (c *UserMgmtClient) PerformGetUser(ctx context.Context, token string, userId string) (*user_mgmt.UserResponse, error) {
	return c.UserMgmtClient.GetUser(ctx, &user_mgmt.GetUserRequest{Token: token, UserId: userId})
}
