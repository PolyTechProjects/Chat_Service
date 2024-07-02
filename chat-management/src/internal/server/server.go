package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	chatMgmt "example.com/chat-management/src/gen/go/chat_mgmt"
	"example.com/chat-management/src/internal/client"
	"example.com/chat-management/src/internal/controller"
	"example.com/chat-management/src/internal/dto"
	"example.com/chat-management/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	chatMgmtController *controller.ChatManagementController
}

func NewHttpServer(chatMgmtController *controller.ChatManagementController) *HttpServer {
	return &HttpServer{chatMgmtController: chatMgmtController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /chat", h.chatMgmtController.CreateChatHandler)
	http.HandleFunc("DELETE /chat", h.chatMgmtController.DeleteChatHandler)
	http.HandleFunc("PUT /chat", h.chatMgmtController.UpdateChatHandler)
	http.HandleFunc("POST /join", h.chatMgmtController.JoinChatHandler)
	http.HandleFunc("POST /leave", h.chatMgmtController.LeaveChatHandler)
	http.HandleFunc("POST /invite", h.chatMgmtController.InviteUserHandler)
	http.HandleFunc("POST /kick", h.chatMgmtController.KickUserHandler)
	http.HandleFunc("POST /admin", h.chatMgmtController.MakeAdminHandler)
	http.HandleFunc("DELETE /admin", h.chatMgmtController.DeleteAdminHandler)
	http.HandleFunc("GET /chat", h.chatMgmtController.GetChatHandler)
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	chatMgmt.UnimplementedChatManagementServer
	service    *service.ChatManagementService
	authClient *client.AuthGRPCClient
}

func New(service *service.ChatManagementService, authClient *client.AuthGRPCClient) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer: gRPCServer,
		service:    service,
		authClient: authClient,
	}
	chatMgmt.RegisterChatManagementServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) CreateChat(ctx context.Context, req *chatMgmt.CreateChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Create chat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.CreatorId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	creatorId, err := uuid.Parse(req.CreatorId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	createReq := &dto.CreateChatRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatorId:   creatorId,
	}
	chatResponse, err := s.service.CreateChat(createReq)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info(fmt.Sprintf("Chat Room %v created successfully", chatResponse.ChatId))

	getReq := &dto.GetChatRequest{
		ChatId: uuid.MustParse(chatResponse.ChatId),
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) DeleteChat(ctx context.Context, req *chatMgmt.DeleteChatRequest) (*chatMgmt.DeleteChatResponse, error) {
	slog.Info("Delete chat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	delReq := &dto.DeleteChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.DeleteChat(delReq)
	if err != nil {
		slog.Error("DeleteChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete chat controller successful", "chatId", req.ChatId)
	return &chatMgmt.DeleteChatResponse{}, nil
}

func (s *GRPCServer) UpdateChat(ctx context.Context, req *chatMgmt.UpdateChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Update chat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	updateReq := &dto.UpdateChatRequest{
		ChatId:      chatId,
		Name:        req.Name,
		Description: req.Description,
		UserId:      userId,
	}
	err = s.service.UpdateChat(updateReq)
	if err != nil {
		slog.Error("UpdateChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update chat controller successful", "chatId", req.ChatId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) JoinChat(ctx context.Context, req *chatMgmt.JoinChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Join chat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	joinReq := &dto.JoinChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.JoinChat(joinReq)
	if err != nil {
		slog.Error("JoinChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join chat controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) LeaveChat(ctx context.Context, req *chatMgmt.LeaveChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Leave chat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	leaveReq := &dto.LeaveChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.LeaveChat(leaveReq)
	if err != nil {
		slog.Error("LeaveChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Leave chat controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) InviteUser(ctx context.Context, req *chatMgmt.InviteUserRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Invite user controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	reqUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.IsAdminRequest{
		ChatId: chatId,
		UserId: userId,
	}
	if isAdmin, err := s.service.IsAdmin(adminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return nil, status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}

	inviteReq := &dto.InviteUserRequest{
		ChatId: chatId,
		UserId: reqUserId,
	}
	err = s.service.InviteUser(inviteReq)
	if err != nil {
		slog.Error("InviteUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Invite user controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *chatMgmt.KickUserRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Kick user controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	reqUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.IsAdminRequest{
		ChatId: chatId,
		UserId: userId,
	}
	if isAdmin, err := s.service.IsAdmin(adminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return nil, status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}

	kickReq := &dto.KickUserRequest{
		ChatId: chatId,
		UserId: reqUserId,
	}
	err = s.service.KickUser(kickReq)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *chatMgmt.MakeAdminRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("MakeAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	reqUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid requesting user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.AdminRequest{
		ChatId:           chatId,
		UserId:           userId,
		RequestingUserId: reqUserId,
	}
	err = s.service.MakeAdmin(adminReq)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChatWithAdmins(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
		ParticipantsIds: response.Users,
		AdminsIds:       response.Admins,
	}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *chatMgmt.DeleteAdminRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("DeleteAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	reqUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid requesting user ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.AdminRequest{
		ChatId:           chatId,
		UserId:           userId,
		RequestingUserId: reqUserId,
	}
	err = s.service.DeleteAdmin(adminReq)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChatWithAdmins(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId.String(),
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
		ParticipantsIds: response.Users,
		AdminsIds:       response.Admins,
	}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *chatMgmt.IsAdminRequest) (*chatMgmt.IsAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.IsAdminRequest{
		ChatId: chatId,
		UserId: userId,
	}
	isAdmin, err := s.service.IsAdmin(adminReq)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)
	return &chatMgmt.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChat(ctx context.Context, req *chatMgmt.GetChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("GetChat controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	chat, err := s.service.GetChat(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          chat.Chat.Id.String(),
		CreatorId:       chat.Chat.CreatorId.String(),
		Name:            chat.Chat.Name,
		Description:     chat.Chat.Description,
		ParticipantsIds: chat.Users,
		AdminsIds:       chat.Admins,
	}, nil
}
