package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-management/src/config"
	"example.com/chat-management/src/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AuthGRPCClient struct {
	auth.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.Dial(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthGRPCClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(ctx context.Context) (*auth.AuthorizeResponse, error) {
	accessToken := metadata.ValueFromIncomingContext(ctx, "authorization")[0]
	refreshToken := metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]

	return authClient.Authorize(ctx, &auth.AuthorizeRequest{AccessToken: accessToken, RefreshToken: refreshToken})
}
