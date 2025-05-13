package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/chat/src/gen/go/chat"
	"example.com/chat/src/internal/controller"
	"example.com/chat/src/internal/models"
	"example.com/chat/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	chatController *controller.ChatController
}

func NewHttpServer(chatController *controller.ChatController) *HttpServer {
	return &HttpServer{chatController: chatController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("GET /api/v1/chats/{chatId}", h.chatController.GetChatHandler)
	http.HandleFunc("DELETE /api/v1/chats/{chatId}", h.chatController.DeleteChatHandler)
	http.HandleFunc("POST /api/v1/chats", h.chatController.CreateChatHandler)
	http.HandleFunc("PUT /api/v1/chats/{chatId}", h.chatController.EditChatHandler)
	http.HandleFunc("PUT /api/v1/chats/join/{joinLink}", h.chatController.JoinChatHandler)
	http.HandleFunc("POST /api/v1/chats/{chatId}/users", h.chatController.AddUsersInChatHandler)
	http.HandleFunc("DELETE /api/v1/chats/{chatId}/users", h.chatController.DeleteUsersInChatHandler)
	http.HandleFunc("POST /api/v1/chats/{chatId}/roles", h.chatController.CreateRoleHandler)
	http.HandleFunc("PUT /api/v1/chats/{chatId}/roles/{roleId}", h.chatController.EditRoleHandler)
	http.HandleFunc("DELETE /api/v1/chats/{chatId}/roles/{roleId}", h.chatController.DeleteRoleHandler)
	http.HandleFunc("PUT /api/v1/chats/{chatId}/users/{userId}/role", h.chatController.SetRoleHandler)
	http.HandleFunc("PATCH /api/v1/chats/{chatId}/users/{userId}/nickname", h.chatController.ChangeUserNicknameHandler)
	http.HandleFunc("GET /api/v1/chats/direct/{userId}", h.chatController.GetDirectChatHandler)
	http.HandleFunc("POST /api/v1/chats/direct", h.chatController.CreateDirectChatHandler)
	http.HandleFunc("DELETE /api/v1/chats", h.chatController.LeaveFromChatsHandler)
	http.HandleFunc("GET /api/v1/chats", h.chatController.GetAllAvailableChatsHandler)
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	chat.UnimplementedChatServer
	chatService *service.ChatService
}

func NewGrpcServer(chatService *service.ChatService) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:  gRPCServer,
		chatService: chatService,
	}
	chat.RegisterChatServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) GetChat(ctx context.Context, req *chat.GetChatRequest) (*chat.ChatResponse, error) {
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
	chatEntity, err := s.chatService.GetChat(chatId, userId)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	users := make([]string, len(chatEntity.Participants))
	for i, user := range chatEntity.Participants {
		users[i] = user.UserId
	}
	return &chat.ChatResponse{
		ChatId:          chatEntity.Chat.Id.String(),
		Name:            chatEntity.Chat.Name,
		Description:     chatEntity.Chat.Description,
		ProfilePic:      chatEntity.Chat.ProfilePic,
		IsChannel:       chatEntity.Chat.IsChannel,
		IsClosed:        chatEntity.Chat.IsClosed,
		JoinLink:        chatEntity.Chat.JoinLink,
		CreatorId:       chatEntity.Chat.CreatorId.String(),
		ParticipantsIds: users,
	}, nil
}

func (s *GRPCServer) VerifyUserAction(ctx context.Context, req *chat.VerifyUserActionRequest) (*chat.VerifyUserActionResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("VerifyUserAction: UUIDParse failed: "+err.Error(), "chat_id", req.ChatId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("VerifyUserAction: UUIDParse failed: "+err.Error(), "user_id", req.UserId)
		return nil, err
	}
	action := models.Permission(req.Action)
	err = s.chatService.VerifyUserAction(chatId, userId, action)
	if err != nil {
		slog.Error("VerifyUserAction failed: " + err.Error())
		if errors.Is(err, fmt.Errorf("permission denied")) {
			return &chat.VerifyUserActionResponse{
				IsVerified: false,
			}, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &chat.VerifyUserActionResponse{
		IsVerified: true,
	}, nil
}

func (s *GRPCServer) VerifyUserPersistance(ctx context.Context, req *chat.VerifyUserPersistanceRequest) (*chat.VerifyUserPersistanceResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("VerifyUserPersistance: UUIDParse failed: "+err.Error(), "chat_id", req.ChatId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("VerifyUserPersistance: UUIDParse failed: "+err.Error(), "user_id", req.UserId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.chatService.VerifyUserPersistance(chatId, userId)
	if err != nil {
		slog.Error("VerifyUserPersistance failed: " + err.Error())
		if errors.Is(err, fmt.Errorf("permission denied")) {
			return &chat.VerifyUserPersistanceResponse{
				IsVerified: false,
			}, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &chat.VerifyUserPersistanceResponse{
		IsVerified: true,
	}, nil
}

func (s *GRPCServer) GetChatAndUserNames(ctx context.Context, req *chat.GetChatAndUserNamesRequest) (*chat.GetChatAndUserNamesResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("GetChatAndUserNames: UUIDParse failed: "+err.Error(), "chat_id", req.ChatId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("GetChatAndUserNames: UUIDParse failed: "+err.Error(), "user_id", req.UserId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	chatName, userName, err := s.chatService.GetChatAndUserNames(chatId, userId)
	if err != nil {
		slog.Error("GetChatAndUserNames error", "error", err.Error())
		return nil, err
	}
	return &chat.GetChatAndUserNamesResponse{
		ChatName: chatName,
		UserName: userName,
	}, nil
}

func (s *GRPCServer) VerifyUserActionOnSomebody(
	ctx context.Context,
	req *chat.VerifyUserActionOnSomebodyRequest,
) (*chat.VerifyUserActionOnSomebodyResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("VerifyUserAction: UUIDParse failed: "+err.Error(), "chat_id", req.ChatId)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("VerifyUserAction: UUIDParse failed: "+err.Error(), "user_id", req.UserId)
		return nil, err
	}
	targetUserId, err := uuid.Parse(req.TargetUserId)
	if err != nil {
		slog.Error("VerifyUserAction: UUIDParse failed: "+err.Error(), "target_user_id", req.TargetUserId)
		return nil, err
	}
	action := models.Permission(req.Action)
	err = s.chatService.VerifyUserActionOnSomebody(chatId, userId, targetUserId, action)
	if err != nil {
		slog.Error("VerifyUserAction failed: " + err.Error())
		if errors.Is(err, fmt.Errorf("permission denied")) {
			return &chat.VerifyUserActionOnSomebodyResponse{
				IsVerified: false,
			}, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &chat.VerifyUserActionOnSomebodyResponse{
		IsVerified: true,
	}, nil
}
