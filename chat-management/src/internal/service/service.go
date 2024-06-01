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

func (s *ChatManagementService) CreateChat(ctx context.Context, req dto.CreateChatRequest) (dto.ChatResponse, error) {
	slog.Info("CreateChat called", "name", req.Name, "description", req.Description, "creatorID", req.CreatorID)
	chat := models.Chat{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   req.CreatorID,
	}
	err := s.repo.SaveChat(chat)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(req.CreatorID)
	if err != nil {
		slog.Error("Invalid creator ID", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddUserToChat(dto.UserChatRequest{ChatID: chat.ID, UserID: userID})
	if err != nil {
		slog.Error("Failed to add user to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddAdmin(dto.AdminRequest{ChatID: chat.ID, UserID: userID})
	if err != nil {
		slog.Error("Failed to add admin to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("Chat created successfully", "chatID", chat.ID)
	return dto.ChatResponse{ChatID: chat.ID.String()}, nil
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

func (s *ChatManagementService) UpdateChat(ctx context.Context, req dto.UpdateChatRequest, userID string) error {
	slog.Info("UpdateChat called", "chatID", req.ChatID, "name", req.Name, "description", req.Description, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, req.ChatID.String(), userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.UpdateChat(req)
	if err != nil {
		slog.Error("Failed to update chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Chat updated successfully", "chatID", req.ChatID)
	return nil
}

func (s *ChatManagementService) JoinChat(ctx context.Context, req dto.UserChatRequest) error {
	slog.Info("JoinChat called", "chatID", req.ChatID, "userID", req.UserID)
	err := s.repo.AddUserToChat(req)
	if err != nil {
		slog.Error("Failed to join chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User joined chat successfully", "chatID", req.ChatID, "userID", req.UserID)
	return nil
}

func (s *ChatManagementService) KickUser(ctx context.Context, chatID, userID, requestingUserID string) error {
	slog.Info("KickUser called", "chatID", chatID, "userID", userID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, chatID, requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.RemoveUserFromChat(dto.UserChatRequest{ChatID: chatUUID, UserID: userUUID})
	if err != nil {
		slog.Error("Failed to kick user from chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User kicked from chat successfully", "chatID", chatID, "userID", userID)
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
	isMember, err := s.repo.IsMember(dto.UserChatRequest{ChatID: chatUUID, UserID: userUUID})
	if err != nil {
		slog.Error("Failed to check member status", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return isMember, nil
}

func (s *ChatManagementService) MakeAdmin(ctx context.Context, req dto.AdminRequest, requestingUserID string) error {
	slog.Info("MakeAdmin called", "chatID", req.ChatID, "userID", req.UserID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, req.ChatID.String(), requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.AddAdmin(req)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin added successfully", "chatID", req.ChatID, "userID", req.UserID)
	return nil
}

func (s *ChatManagementService) DeleteAdmin(ctx context.Context, req dto.AdminRequest, requestingUserID string) error {
	slog.Info("DeleteAdmin called", "chatID", req.ChatID, "userID", req.UserID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, req.ChatID.String(), requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.RemoveAdmin(req)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin removed successfully", "chatID", req.ChatID, "userID", req.UserID)
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
	isAdmin, err := s.repo.IsAdmin(dto.AdminRequest{ChatID: chatUUID, UserID: userUUID})
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
