package server

import (
	"context"
	"log/slog"
	"net"

	chatMgmt "example.com/chat-management/src/gen/go/chat-mgmt"
	"example.com/chat-management/src/internal/client"
	"example.com/chat-management/src/internal/dto"
	"example.com/chat-management/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	chatMgmt.UnimplementedChatManagementServer
	service    *service.ChatManagementService
	authClient *client.AuthGRPCClient
}

func New(service *service.ChatManagementService, authClient *client.AuthGRPCClient) (*GRPCServer, error) {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer: gRPCServer,
		service:    service,
		authClient: authClient,
	}
	chatMgmt.RegisterChatManagementServer(gRPCServer, g)
	return g, nil
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) CreateChat(ctx context.Context, req *chatMgmt.CreateChatRequest) (*chatMgmt.CreateChatResponse, error) {
	slog.Info("Create chat controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	createReq := dto.CreateChatRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   req.CreatorId,
	}
	chatResponse, err := s.service.CreateChat(ctx, createReq)
	if err != nil {
		slog.Error("CreateChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Create chat controller successful", "chatID", chatResponse.ChatID)
	return &chatMgmt.CreateChatResponse{ChatId: chatResponse.ChatID}, nil
}

func (s *GRPCServer) DeleteChat(ctx context.Context, req *chatMgmt.DeleteChatRequest) (*chatMgmt.DeleteChatResponse, error) {
	slog.Info("Delete chat controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.DeleteChat(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("DeleteChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete chat controller successful", "chatID", req.ChatId)
	return &chatMgmt.DeleteChatResponse{}, nil
}

func (s *GRPCServer) UpdateChat(ctx context.Context, req *chatMgmt.UpdateChatRequest) (*chatMgmt.UpdateChatResponse, error) {
	slog.Info("Update chat controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	updateReq := dto.UpdateChatRequest{
		ChatID:      chatID,
		Name:        req.Name,
		Description: req.Description,
	}
	err = s.service.UpdateChat(ctx, updateReq, req.UserId)
	if err != nil {
		slog.Error("UpdateChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update chat controller successful", "chatID", req.ChatId)
	return &chatMgmt.UpdateChatResponse{}, nil
}

func (s *GRPCServer) JoinChat(ctx context.Context, req *chatMgmt.JoinChatRequest) (*chatMgmt.JoinChatResponse, error) {
	slog.Info("Join chat controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	joinReq := dto.UserChatRequest{
		ChatID: chatID,
		UserID: userID,
	}
	err = s.service.JoinChat(ctx, joinReq)
	if err != nil {
		slog.Error("JoinChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join chat controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.JoinChatResponse{}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *chatMgmt.KickUserChatRequest) (*chatMgmt.KickUserChatResponse, error) {
	slog.Info("Kick user controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.KickUser(ctx, req.ChatId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.KickUserChatResponse{}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *chatMgmt.CanWriteChatRequest) (*chatMgmt.CanWriteChatResponse, error) {
	slog.Info("CanWrite controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	canWrite, err := s.service.CanWrite(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, err
	}
	slog.Info("CanWrite controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.CanWriteChatResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *chatMgmt.MakeChatAdminRequest) (*chatMgmt.MakeChatAdminResponse, error) {
	slog.Info("MakeAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := dto.AdminRequest{
		ChatID: chatID,
		UserID: userID,
	}
	err = s.service.MakeAdmin(ctx, adminReq, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.MakeChatAdminResponse{}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *chatMgmt.DeleteChatAdminRequest) (*chatMgmt.DeleteChatAdminResponse, error) {
	slog.Info("DeleteAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	chatID, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := dto.AdminRequest{
		ChatID: chatID,
		UserID: userID,
	}
	err = s.service.DeleteAdmin(ctx, adminReq, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.DeleteChatAdminResponse{}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *chatMgmt.IsChatAdminRequest) (*chatMgmt.IsChatAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	isAdmin, err := s.service.IsAdmin(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.IsChatAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChatUsers(ctx context.Context, req *chatMgmt.GetChatUsersRequest) (*chatMgmt.GetChatUsersResponse, error) {
	slog.Info("GetChatUsers controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	userIDs, err := s.service.GetChatUsers(ctx, req.ChatId)
	if err != nil {
		slog.Error("GetChatUsers error", "error", err.Error())
		return nil, err
	}
	slog.Info("GetChatUsers controller successful", "chatID", req.ChatId)
	return &chatMgmt.GetChatUsersResponse{UserIds: userIDs}, nil
}
