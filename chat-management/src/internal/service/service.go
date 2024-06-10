package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/chat-management/src/internal/dto"
	"example.com/chat-management/src/internal/models"
	"example.com/chat-management/src/internal/repository"
	"github.com/google/uuid"
)

type ChatManagementService struct {
	repo repository.ChatRepository
}

func New(repo repository.ChatRepository) *ChatManagementService {
	return &ChatManagementService{
		repo: repo,
	}
}

func (s *ChatManagementService) GetChat(ctx context.Context, req *dto.GetChatRequest) (*dto.GetChatResponse, error) {
	slog.Info("GetChat called", "chatID", req.ChatId)
	chat, err := s.repo.FindById(req.ChatId)
	if err != nil {
		slog.Error("Failed to get chat", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	users, err := s.repo.GetChatUsers(req.ChatId)
	if err != nil {
		slog.Error("Failed to get chat users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getResp := &dto.GetChatResponse{
		Chat:  chat,
		Users: users,
	}
	return getResp, nil
}

func (s *ChatManagementService) GetChatWithAdmins(ctx context.Context, req *dto.GetChatRequest) (*dto.GetChatResponse, error) {
	getResp, err := s.GetChat(ctx, req)
	if err != nil {
		return nil, err
	}
	admins, err := s.repo.GetChatAdmins(req.ChatId)
	if err != nil {
		slog.Error("Failed to get chat users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getResp.Admins = admins
	return getResp, nil
}

func (s *ChatManagementService) CreateChat(ctx context.Context, req *dto.CreateChatRequest) (dto.ChatResponse, error) {
	slog.Info("CreateChat called", "name", req.Name, "description", req.Description, "creatorID", req.CreatorId)
	chat := models.Chat{
		Id:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatorId:   req.CreatorId,
	}
	err := s.repo.SaveChat(chat)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.CreatorId)
	if err != nil {
		slog.Error("Invalid creator ID", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddUserToChat(&dto.UserChatRequest{ChatId: chat.Id, UserId: userID})
	if err != nil {
		slog.Error("Failed to add user to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddAdmin(&dto.AdminRequest{ChatId: chat.Id, UserId: userID})
	if err != nil {
		slog.Error("Failed to add admin to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("Chat created successfully", "chatID", chat.Id)
	return dto.ChatResponse{ChatId: chat.Id.String()}, nil
}

func (s *ChatManagementService) DeleteChat(ctx context.Context, chatID, userID string) error {
	slog.Info("DeleteChat called", "chatID", chatID, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, chatID, userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	id, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.DeleteChat(id)
	if err != nil {
		slog.Error("Failed to delete chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Chat deleted successfully", "chatID", chatID)
	return nil
}

func (s *ChatManagementService) UpdateChat(ctx context.Context, req *dto.UpdateChatRequest, userID string) error {
	slog.Info("UpdateChat called", "chatID", req.ChatId, "name", req.Name, "description", req.Description, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, req.ChatId.String(), userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.UpdateChat(req)
	if err != nil {
		slog.Error("Failed to update chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Chat updated successfully", "chatID", req.ChatId)
	return nil
}

func (s *ChatManagementService) AddUser(ctx context.Context, req *dto.UserChatRequest) error {
	slog.Info("JoinChat called", "chatID", req.ChatId, "userID", req.UserId)
	err := s.repo.AddUserToChat(req)
	if err != nil {
		slog.Error("Failed to join chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User joined chat successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) RemoveUser(ctx context.Context, req *dto.UserChatRequest) error {
	slog.Info("LeaveChat called", "chatID", req.ChatId, "userID", req.UserId)
	err := s.repo.RemoveUserFromChat(req)
	if err != nil {
		slog.Error("Failed to leave chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User left chat successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) CanWrite(ctx context.Context, chatID, userID string) (bool, error) {
	slog.Info("CanWrite called", "chatID", chatID, "userID", userID)
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	isMember, err := s.repo.IsMember(&dto.UserChatRequest{ChatId: chatUUID, UserId: userUUID})
	if err != nil {
		slog.Error("Failed to check member status", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return isMember, nil
}

func (s *ChatManagementService) MakeAdmin(ctx context.Context, req *dto.AdminRequest, requestingUserId string) error {
	slog.Info("MakeAdmin called", "chatID", req.ChatId, "userID", req.UserId, "requestingUserId", requestingUserId)
	if isAdmin, err := s.isAdmin(ctx, req.ChatId.String(), requestingUserId); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.AddAdmin(req)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin added successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) DeleteAdmin(ctx context.Context, req *dto.AdminRequest, requestingUserId string) error {
	slog.Info("DeleteAdmin called", "chatID", req.ChatId, "userID", req.UserId, "requestingUserId", requestingUserId)
	if isAdmin, err := s.isAdmin(ctx, req.ChatId.String(), requestingUserId); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.RemoveAdmin(req)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin removed successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) IsAdmin(ctx context.Context, chatID, userID string) (bool, error) {
	slog.Info("IsAdmin called", "chatID", chatID, "userID", userID)
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	isAdmin, err := s.repo.IsAdmin(&dto.AdminRequest{ChatId: chatUUID, UserId: userUUID})
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return isAdmin, nil
}

func (s *ChatManagementService) isAdmin(ctx context.Context, chatID, userID string) (bool, error) {
	return s.IsAdmin(ctx, chatID, userID)
}

func (s *ChatManagementService) GetChatUsers(ctx context.Context, chatID string) ([]string, error) {
	slog.Info("GetChatUsers called", "chatID", chatID)
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userIDs, err := s.repo.GetChatUsers(chatUUID)
	if err != nil {
		slog.Error("Failed to get chat users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return userIDs, nil
}
