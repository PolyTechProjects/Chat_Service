package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/notification/src/config"
	"example.com/notification/src/gen/go/auth"
	userMgmt "example.com/notification/src/gen/go/user_mgmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	auth.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthClient) PerformAuthorize(ctx context.Context, accessToken string, refreshToken string) (*auth.AuthorizeResponse, error) {
	return authClient.Authorize(ctx, &auth.AuthorizeRequest{AccessToken: accessToken, RefreshToken: refreshToken})
}

type UserMgmtClient struct {
	userMgmt.UserMgmtClient
}

func NewUserMgmtClient(cfg *config.Config) *UserMgmtClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.UserMgmt.UserMgmtHost, cfg.UserMgmt.UserMgmtPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return &UserMgmtClient{userMgmt.NewUserMgmtClient(conn)}
}

func (c *UserMgmtClient) PerformGetUser(ctx context.Context, token string, userId string) (*userMgmt.UserResponse, error) {
	return c.UserMgmtClient.GetUser(ctx, &userMgmt.GetUserRequest{Token: token, UserId: userId})
}
