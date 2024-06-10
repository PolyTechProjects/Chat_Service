package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/gen/go/auth"
	"example.com/user-mgmt/src/gen/go/media"
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
	slog.Info("Connected to Auth")
	slog.Info(connectionUrl)
	return &AuthGRPCClient{auth.NewAuthClient(conn)}
}

func (authClient *AuthGRPCClient) PerformAuthorize(ctx context.Context, r *http.Request) (*auth.AuthorizeResponse, error) {
	var accessToken, refreshToken string
	if r == nil {
		accessToken = metadata.ValueFromIncomingContext(ctx, "authorization")[0]
		refreshToken = metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]
	} else {
		ctx = r.Context()
		authHeader := r.Header.Get("Authorization")
		accessToken = strings.Split(authHeader, " ")[1]
		cookie, err := r.Cookie("X-Refresh-Token")
		if err != nil {
			return nil, err
		}
		refreshToken = cookie.Value
	}
	return authClient.Authorize(ctx, &auth.AuthorizeRequest{AccessToken: accessToken, RefreshToken: refreshToken})
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
