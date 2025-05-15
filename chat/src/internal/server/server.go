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

func (h *HttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/chats/{chatId}", h.chatController.GetChatHandler)
	mux.HandleFunc("DELETE /api/v1/chats/{chatId}", h.chatController.DeleteChatHandler)
	mux.HandleFunc("POST /api/v1/chats", h.chatController.CreateChatHandler)
	mux.HandleFunc("PUT /api/v1/chats/{chatId}", h.chatController.EditChatHandler)
	mux.HandleFunc("PUT /api/v1/chats/join/{joinLink}", h.chatController.JoinChatHandler)
	mux.HandleFunc("POST /api/v1/chats/{chatId}/users", h.chatController.AddUsersInChatHandler)
	mux.HandleFunc("DELETE /api/v1/chats/{chatId}/users", h.chatController.DeleteUsersInChatHandler)
	mux.HandleFunc("POST /api/v1/chats/{chatId}/roles", h.chatController.CreateRoleHandler)
	mux.HandleFunc("PUT /api/v1/chats/{chatId}/roles/{roleId}", h.chatController.EditRoleHandler)
	mux.HandleFunc("DELETE /api/v1/chats/{chatId}/roles/{roleId}", h.chatController.DeleteRoleHandler)
	mux.HandleFunc("PUT /api/v1/chats/{chatId}/users/{userId}/role", h.chatController.SetRoleHandler)
	mux.HandleFunc("GET /api/v1/chats/{chatId}/roles", h.chatController.GetChatRolesHandler)
	mux.HandleFunc("GET /api/v1/chats/{chatId}/roles/{roleId}", h.chatController.GetChatRoleHandler)
	mux.HandleFunc("GET /api/v1/chats/{chatId}/permissions", h.chatController.GetAllPermissionsHandler)
	mux.HandleFunc("PATCH /api/v1/chats/{chatId}/users/{userId}/nickname", h.chatController.ChangeUserNicknameHandler)
	//mux.HandleFunc("GET /api/v1/chats/direct/{userId}", h.chatController.GetDirectChatHandler)
	mux.HandleFunc("GET /api/v1/chats/{chatId}/users/{userId}", h.chatController.GetChatUserHandler)
	mux.HandleFunc("POST /api/v1/chats/direct", h.chatController.CreateDirectChatHandler)
	mux.HandleFunc("DELETE /api/v1/chats", h.chatController.LeaveFromChatsHandler)
	mux.HandleFunc("GET /api/v1/chats", h.chatController.GetAllAvailableChatsHandler)
}

func (h *HttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
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

func (s *GRPCServer) GetDirectChat(ctx context.Context, req *chat.GetDirectChatRequest) (*chat.DirectChatResponse, error) {
	chatId, err := uuid.Parse(req.ChatId)
	if err != nil {
		slog.Error("Invalid chat Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.OwnUserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	chatEntity, err := s.chatService.GetDirectChatById(chatId, userId)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	if chatEntity.FirstUserId != userId {
		chatEntity.FirstUserId, chatEntity.SecondUserId = chatEntity.SecondUserId, chatEntity.FirstUserId
	}
	return &chat.DirectChatResponse{
		ChatId:       chatEntity.ChatId.String(),
		OwnUserId:    chatEntity.FirstUserId.String(),
		TargetUserId: chatEntity.SecondUserId.String(),
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
