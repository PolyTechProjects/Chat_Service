package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"example.com/notification/src/gen/go/notification"
	userMgmt "example.com/notification/src/gen/go/user_mgmt"
	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/service"
	"example.com/notification/src/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type NotificationServer struct {
	gRPCServer *grpc.Server
	notification.UnimplementedNotificationServer
	notificationService *service.NotificationService
	authClient          *client.AuthClient
	userMgmtClient      *client.UserMgmtClient
}

func NewNotificationServer(notificationService *service.NotificationService, userMgmtClient *client.UserMgmtClient) *NotificationServer {
	gRPCServer := grpc.NewServer()
	g := &NotificationServer{
		gRPCServer:          gRPCServer,
		notificationService: notificationService,
		userMgmtClient:      userMgmtClient,
	}
	notification.RegisterNotificationServer(gRPCServer, g)
	return g
}

func (s *NotificationServer) Start(l net.Listener) error {
	slog.Debug("Starting notification gRPC server")
	slog.Debug(l.Addr().String())
	slog.Debug("Listening notification channel")
	go s.ListenNotificationChannel()
	return s.gRPCServer.Serve(l)
}

func (s *NotificationServer) ListenNotificationChannel() {
	var subscriber = s.notificationService.SubscribeToRedisChannel("notification-channel")
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Not Available notification-channel")
	}
	slog.Info("Available notification-channel")
	for {
		channel := subscriber.Channel()
		message := <-channel
		slog.Info(message.Payload)
		readyMessage := &models.ReadyMessage{}
		err = json.Unmarshal([]byte(message.Payload), readyMessage)
		if err != nil {
			slog.Error(err.Error())
			break
		}
		user, err := s.userMgmtClient.GetUser(context.Background(), &userMgmt.GetUserRequest{UserId: readyMessage.Message.SenderId.String()})
		if err != nil {
			slog.Error(err.Error())
			break
		}
		err = s.notificationService.NotifyUsers(
			readyMessage.ReceiversIds,
			readyMessage.Message.CreatedAt,
			readyMessage.Message.Body,
			user.Name,
			user.Avatar,
		)
		if err != nil {
			slog.Error(err.Error())
			break
		}
	}
}

func (s *NotificationServer) BindDeviceToUser(ctx context.Context, req *notification.BindDeviceRequest) (*notification.BindDeviceResponse, error) {
	accessToken := metadata.ValueFromIncomingContext(ctx, "authorization")[0]
	refreshToken := metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]
	authResp, err := s.authClient.PerformAuthorize(ctx, accessToken, refreshToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userUUID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.notificationService.BindDeviceToUser(userUUID, req.GetDeviceToken())
	return &notification.BindDeviceResponse{}, nil
}

func (s *NotificationServer) UnbindDeviceFromUser(ctx context.Context, req *notification.UnbindDeviceRequest) (*notification.UnbindDeviceResponse, error) {
	accessToken := metadata.ValueFromIncomingContext(ctx, "authorization")[0]
	refreshToken := metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]
	authResp, err := s.authClient.PerformAuthorize(ctx, accessToken, refreshToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userUUID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.notificationService.UnbindDeviceFromUser(userUUID, req.GetDeviceToken())
	return &notification.UnbindDeviceResponse{}, nil
}

func (s *NotificationServer) DeleteUser(ctx context.Context, req *notification.DeleteUserRequest) (*notification.DeleteUserResponse, error) {
	accessToken := metadata.ValueFromIncomingContext(ctx, "authorization")[0]
	refreshToken := metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]
	authResp, err := s.authClient.PerformAuthorize(ctx, accessToken, refreshToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userUUID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.notificationService.DeleteUser(userUUID)
	return &notification.DeleteUserResponse{}, nil
}

func (s *NotificationServer) UpdateOldDeviceOnUser(ctx context.Context, req *notification.UpdateOldDeviceRequest) (*notification.UpdateOldDeviceResponse, error) {
	accessToken := metadata.ValueFromIncomingContext(ctx, "authorization")[0]
	refreshToken := metadata.ValueFromIncomingContext(ctx, "x-refresh-token")[0]
	authResp, err := s.authClient.PerformAuthorize(ctx, accessToken, refreshToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userUUID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.notificationService.UpdateOldDeviceOnUser(userUUID, req.GetOldDeviceToken(), req.GetNewDeviceToken())
	return &notification.UpdateOldDeviceResponse{}, nil
}
