package client

import (
	"context"
	"fmt"
	"log/slog"

	"example.com/chat-management/src/config"
	"example.com/chat-management/src/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthGRPCClient struct {
	sso.AuthClient
}

func NewAuthClient(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.Dial(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthGRPCClient{sso.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Error("Missing metadata")
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	token := md["authorization"]
	if len(token) == 0 {
		slog.Error("Missing token")
		return status.Error(codes.Unauthenticated, "missing token")
	}

	req := &sso.AuthorizeRequest{Token: token[0]}
	res, err := authClient.Authorize(ctx, req)
	if err != nil {
		slog.Error("Authorization failed", "error", err.Error())
		return status.Error(codes.PermissionDenied, "unauthorized")
	}
	if !res.Authorized {
		slog.Error("Unauthorized access attempt")
		return status.Error(codes.PermissionDenied, "unauthorized")
	}
	return nil
}

func (authClient *AuthGRPCClient) PerformUserIdExtraction(ctx context.Context, token string) (*sso.ExtractUserIdResponse, error) {
	return authClient.ExtractUserId(ctx, &sso.ExtractUserIdRequest{Token: token})
}
