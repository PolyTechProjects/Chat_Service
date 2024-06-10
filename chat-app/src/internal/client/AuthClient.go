package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/config"
	"example.com/chat-app/src/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthGRPCClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(ctx context.Context, accessToken string, refreshToken string) (*auth.AuthorizeResponse, error) {
	return authClient.Authorize(ctx, &auth.AuthorizeRequest{AccessToken: accessToken, RefreshToken: refreshToken})
}
