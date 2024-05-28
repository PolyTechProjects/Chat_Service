package client

import (
	"context"
	"fmt"

	"example.com/media-handler/src/config"
	"example.com/media-handler/src/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCClient struct {
	authClient sso.AuthClient
}

func New(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return &AuthGRPCClient{authClient: sso.NewAuthClient(conn)}
}

func (c *AuthGRPCClient) PerformAuthorize(ctx context.Context, token string) (*sso.AuthorizeResponse, error) {
	return c.authClient.Authorize(ctx, &sso.AuthorizeRequest{Token: token})
}
