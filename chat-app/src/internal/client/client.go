package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-app/src/config"
	"example.com/chat-app/src/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCClient struct {
	sso.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthGRPCClient{sso.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(ctx context.Context, token string) (*sso.AuthorizeResponse, error) {
	return authClient.Authorize(ctx, &sso.AuthorizeRequest{Token: token})
}

func (authClient *AuthGRPCClient) PerformUserIdExtraction(ctx context.Context, token string) (*sso.ExtractUserIdResponse, error) {
	return authClient.ExtractUserId(ctx, &sso.ExtractUserIdRequest{Token: token})
}
