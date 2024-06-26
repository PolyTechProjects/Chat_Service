package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"

	"example.com/notification/src/gen/go/notification"
	userMgmt "example.com/notification/src/gen/go/user_mgmt"
	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/controller"
	"example.com/notification/src/internal/service"
	"example.com/notification/src/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationHttpServer struct {
	notificationController *controller.NotificationController
}

func NewNotificationHttpServer(notificationController *controller.NotificationController) *NotificationHttpServer {
	return &NotificationHttpServer{
		notificationController: notificationController,
	}
}

func (n *NotificationHttpServer) StartServer() {
	http.HandleFunc("POST /device", n.notificationController.BindDeviceToUserHandler)
	http.HandleFunc("DELETE /device", n.notificationController.UnbindDeviceFromUserHandler)
	http.HandleFunc("DELETE /user", n.notificationController.DeleteUserHandler)
	http.HandleFunc("PUT /device", n.notificationController.UpdateOldDeviceOnUserHandler)
}

type NotificationGRPCServer struct {
	gRPCServer *grpc.Server
	notification.UnimplementedNotificationServer
	notificationService *service.NotificationService
	authClient          *client.AuthClient
	userMgmtClient      *client.UserMgmtClient
}

func NewNotificationGRPCServer(notificationService *service.NotificationService, authClient *client.AuthClient, userMgmtClient *client.UserMgmtClient) *NotificationGRPCServer {
	gRPCServer := grpc.NewServer()
	g := &NotificationGRPCServer{
		gRPCServer:          gRPCServer,
		notificationService: notificationService,
		authClient:          authClient,
		userMgmtClient:      userMgmtClient,
	}
	notification.RegisterNotificationServer(gRPCServer, g)
	return g
}

func (s *NotificationGRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting notification gRPC server")
	slog.Debug(l.Addr().String())
	slog.Debug("Listening notification channel")
	go s.ListenNotificationChannel()
	return s.gRPCServer.Serve(l)
}

func (s *NotificationGRPCServer) ListenNotificationChannel() {
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

func (s *NotificationGRPCServer) BindDeviceToUser(ctx context.Context, req *notification.BindDeviceRequest) (*notification.BindDeviceResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
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

func (s *NotificationGRPCServer) UnbindDeviceFromUser(ctx context.Context, req *notification.UnbindDeviceRequest) (*notification.UnbindDeviceResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
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

func (s *NotificationGRPCServer) DeleteUser(ctx context.Context, req *notification.DeleteUserRequest) (*notification.DeleteUserResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
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

func (s *NotificationGRPCServer) UpdateOldDeviceOnUser(ctx context.Context, req *notification.UpdateOldDeviceRequest) (*notification.UpdateOldDeviceResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
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
