package client

import (
	"context"
	"fmt"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCClient struct {
	client sso.AuthClient
}

func New(cfg *config.Config) *AuthGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Auth.AuthHost, cfg.Auth.AuthPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	return &AuthGRPCClient{client: sso.NewAuthClient(conn)}
}

func (c *AuthGRPCClient) PerformAuthorize(ctx context.Context, token string) (*sso.AuthorizeResponse, error) {
	return c.client.Authorize(ctx, &sso.AuthorizeRequest{Token: token})
}
