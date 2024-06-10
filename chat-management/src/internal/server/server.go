package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	chatMgmt "example.com/chat-management/src/gen/go/chat_mgmt"
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

func (s *GRPCServer) CreateChat(ctx context.Context, req *chatMgmt.CreateChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Create chat controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.CreatorId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.CreatorId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	createReq := &dto.CreateChatRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatorId:   req.CreatorId,
	}
	chatResponse, err := s.service.CreateChat(ctx, createReq)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info(fmt.Sprintf("Chat Room %v created successfully", chatResponse.ChatId))

	getReq := &dto.GetChatRequest{
		ChatId: uuid.MustParse(chatResponse.ChatId),
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) DeleteChat(ctx context.Context, req *chatMgmt.DeleteChatRequest) (*chatMgmt.DeleteChatResponse, error) {
	slog.Info("Delete chat controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.DeleteChat(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("DeleteChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete chat controller successful", "chatId", req.ChatId)
	return &chatMgmt.DeleteChatResponse{}, nil
}

func (s *GRPCServer) UpdateChat(ctx context.Context, req *chatMgmt.UpdateChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Update chat controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	updateReq := &dto.UpdateChatRequest{
		ChatId:      chatId,
		Name:        req.Name,
		Description: req.Description,
	}
	err = s.service.UpdateChat(ctx, updateReq, req.UserId)
	if err != nil {
		slog.Error("UpdateChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update chat controller successful", "chatId", req.ChatId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) JoinChat(ctx context.Context, req *chatMgmt.JoinChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Join chat controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
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

	joinReq := &dto.UserChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.AddUser(ctx, joinReq)
	if err != nil {
		slog.Error("JoinChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join chat controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) LeaveChat(ctx context.Context, req *chatMgmt.LeaveChatRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Leave chat controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
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

	leaveReq := &dto.UserChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.RemoveUser(ctx, leaveReq)
	if err != nil {
		slog.Error("LeaveChat error", "error", err.Error())
		return nil, err
	}
	slog.Info("Leave chat controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) InviteUser(ctx context.Context, req *chatMgmt.InviteUserRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Invite user controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
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

	inviteReq := &dto.UserChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.AddUser(ctx, inviteReq)
	if err != nil {
		slog.Error("InviteUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Invite user controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *chatMgmt.KickUserRequest) (*chatMgmt.ChatRoomResponse, error) {
	slog.Info("Kick user controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if isAdmin, err := s.service.IsAdmin(ctx, req.ChatId, req.RequestingUserId); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return nil, status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
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

	kickReq := &dto.UserChatRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.RemoveUser(ctx, kickReq)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChat(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		ParticipantsIds: response.Users,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
	}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *chatMgmt.CanWriteRequest) (*chatMgmt.CanWriteResponse, error) {
	slog.Info("CanWrite controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	canWrite, err := s.service.CanWrite(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, err
	}
	slog.Info("CanWrite controller successful", "chatId", req.ChatId, "userId", req.UserId)
	return &chatMgmt.CanWriteResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *chatMgmt.MakeAdminRequest) (*chatMgmt.ChatRoomWithAdminsResponse, error) {
	slog.Info("MakeAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
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

	adminReq := &dto.AdminRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.MakeAdmin(ctx, adminReq, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChatWithAdmins(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomWithAdminsResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
		ParticipantsIds: response.Users,
		AdminsIds:       response.Admins,
	}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *chatMgmt.DeleteAdminRequest) (*chatMgmt.ChatRoomWithAdminsResponse, error) {
	slog.Info("DeleteAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
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
	adminReq := &dto.AdminRequest{
		ChatId: chatId,
		UserId: userId,
	}
	err = s.service.DeleteAdmin(ctx, adminReq, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)

	getReq := &dto.GetChatRequest{
		ChatId: chatId,
	}
	response, err := s.service.GetChatWithAdmins(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &chatMgmt.ChatRoomWithAdminsResponse{
		ChatId:          response.Chat.Id.String(),
		CreatorId:       response.Chat.CreatorId,
		Name:            response.Chat.Name,
		Description:     response.Chat.Description,
		ParticipantsIds: response.Users,
		AdminsIds:       response.Admins,
	}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *chatMgmt.IsAdminRequest) (*chatMgmt.IsAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	isAdmin, err := s.service.IsAdmin(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "chatId", req.ChatId, "userId", req.UserId)
	return &chatMgmt.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChatUsers(ctx context.Context, req *chatMgmt.GetUsersRequest) (*chatMgmt.GetUsersResponse, error) {
	slog.Info("GetChatUsers controller started")
	_, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	userIds, err := s.service.GetChatUsers(ctx, req.ChatId)
	if err != nil {
		slog.Error("GetChatUsers error", "error", err.Error())
		return nil, err
	}
	slog.Info("GetChatUsers controller successful", "chatId", req.ChatId)
	return &chatMgmt.GetUsersResponse{UserIds: userIds}, nil
}
