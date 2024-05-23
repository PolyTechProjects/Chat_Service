package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/gen/go/media"
	"example.com/user-mgmt/src/gen/go/sso"
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

type MediaHandlerGRPCClient struct {
	media.MediaHandlerClient
}

func NewMediaHandlerClient(cfg *config.Config) *MediaHandlerGRPCClient {
	connectionUrl := fmt.Sprintf("%s:%s", cfg.Media.MediaHandlerHost, cfg.Media.MediaHandlerPort)
	conn, err := grpc.NewClient(connectionUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	slog.Info("Connected to MediaHandler")
	slog.Info(connectionUrl)
	return &MediaHandlerGRPCClient{media.NewMediaHandlerClient(conn)}
}

func (mediaHandlerClient *MediaHandlerGRPCClient) PerformStoreImage(ctx context.Context, token string, file multipart.File, fileName string) (*media.ImageResponse, error) {
	stream, err := mediaHandlerClient.StoreImage(ctx)
	if err != nil {
		return nil, err
	}
	slog.Info("Store Image")
	buf := make([]byte, 1024*16)
	for {
		_, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		slog.Info("Read")
		err = stream.Send(&media.StoreImageRequest{Token: token, Data: buf, FileName: fileName})
		if err != nil {
			return nil, err
		}
		slog.Info("Send")
	}
	return stream.CloseAndRecv()
}
