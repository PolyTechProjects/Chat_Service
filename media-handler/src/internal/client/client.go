package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"

	"example.com/media/src/config"
	"example.com/media/src/gen/go/auth"
	"example.com/media/src/internal/dto"
	"github.com/go-redis/redis/v8"
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

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.InnerPort),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	return &RedisClient{
		Client: client,
	}
}

func (c *RedisClient) CacheVolumeIp(volumeId string, volumeIp string) error {
	_, err := c.Client.Set(context.Background(), fmt.Sprintf("VOLUME_%s", volumeId), volumeIp, 0).Result()
	if err != nil {
		slog.Error("RedisSet failed: " + err.Error())
		return err
	}
	return nil
}

func (c *RedisClient) GetVolumeIp(volumeId string) (string, error) {
	volumeIp, err := c.Client.Get(context.Background(), fmt.Sprintf("VOLUME_%s", volumeId)).Result()
	if err != nil {
		slog.Error("RedisSet failed: " + err.Error())
		return "", err
	}
	return volumeIp, nil
}

type SeaweedFSClient struct {
	masterUrl string
}

func NewSeaweedFSCLient(cfg *config.Config) *SeaweedFSClient {
	return &SeaweedFSClient{masterUrl: fmt.Sprintf("%s:%d", cfg.SeaweedFS.MasterIp, cfg.SeaweedFS.MasterPort)}
}

func (c *SeaweedFSClient) AssignAndUpload(file io.Reader, fileName string) (*dto.SeaweedFSAssignResponse, error) {
	assignResponse := &dto.SeaweedFSAssignResponse{}
	res, err := http.Get(fmt.Sprintf("http://%s/dir/assign", c.masterUrl))
	if err != nil {
		slog.Error("Assign failed: " + err.Error())
		return nil, err
	}
	json.NewDecoder(res.Body).Decode(assignResponse)
	defer res.Body.Close()

	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	form, err := w.CreateFormFile("file", fileName)
	if err != nil {
		slog.Error("httpCreateFormFile failed: " + err.Error())
		return nil, err
	}
	_, err = io.Copy(form, file)
	if err != nil {
		slog.Error("ioCopy failed: " + err.Error())
		return nil, err
	}
	w.Close()

	addr := fmt.Sprintf("http://%s/%s", assignResponse.Url, assignResponse.Fid)
	slog.Info(fmt.Sprintf("File URL: %v", addr))
	req, err := http.NewRequest("POST", addr, b)
	if err != nil {
		slog.Error("httpNewRequest failed: " + err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Upload failed: " + err.Error())
		return nil, err
	}
	return assignResponse, nil
}

func (c *SeaweedFSClient) Lookup(volumeId string) (*dto.SeaweedFSLookupResponse, error) {
	lookupResponse := &dto.SeaweedFSLookupResponse{}
	res, err := http.Get(fmt.Sprintf("http://%s/dir/lookup?volumeId=%s", c.masterUrl, volumeId))
	if err != nil {
		slog.Error("Lookup failed: " + err.Error())
		return nil, err
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(lookupResponse)
	return lookupResponse, nil
}
