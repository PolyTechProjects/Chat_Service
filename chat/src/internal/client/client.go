package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/chat/src/config"
	"example.com/chat/src/gen/go/auth"
	"example.com/chat/src/gen/go/users"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	slog.Info("Connected to Auth: " + connectionUrl)
	return &AuthGRPCClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(r *http.Request) (*auth.AuthorizeResponse, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		slog.Error("PerformAuthorize failed: No access token provided")
		return nil, fmt.Errorf("PerformAuthorize failed: No access token provided")
	}
	ctx := metadata.AppendToOutgoingContext(r.Context(), "Authorization", "Bearer "+accessToken)
	return authClient.Authorize(ctx, &auth.AuthorizeRequest{})
}

func (authClient *AuthGRPCClient) PerformExtractUserId(r *http.Request) (*auth.ExtractUserIdResponse, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		slog.Error("PerformExtractUserId failed: No access token provided")
		return nil, fmt.Errorf("PerformExtractUserId failed: No access token provided")
	}
	ctx := metadata.AppendToOutgoingContext(r.Context(), "Authorization", "Bearer "+accessToken)
	return authClient.ExtractUserId(ctx, &auth.ExtractUserIdRequest{})
}

type UsersGRPCClient struct {
	users.UsersClient
}

func NewUsersClient(cfg *config.Config) *UsersGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Users.UsersHost, cfg.Users.UsersPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to Users: " + connectionUrl)
	return &UsersGRPCClient{users.NewUsersClient(conn)}
}

func (c *UsersGRPCClient) PerformGetUser(userId uuid.UUID) (*users.UserResponse, error) {
	return c.UsersClient.GetUser(context.Background(), &users.GetUserRequest{UserId: userId.String()})
}

func (c *UsersGRPCClient) PerformGetUsers(userUUIds []uuid.UUID) (*users.UsersResponse, error) {
	userIds := make([]string, len(userUUIds))
	for i, id := range userUUIds {
		userIds[i] = id.String()
	}
	return c.UsersClient.GetUsers(context.Background(), &users.GetUsersRequest{UserIds: userIds})
}
