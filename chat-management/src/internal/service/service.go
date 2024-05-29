package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

func (s *ChatManagementService) CreateChat(ctx context.Context, name, description, creatorID string) (string, error) {
	slog.Info("CreateChat called", "name", name, "description", description, "creatorID", creatorID)
	chat := models.Chat{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
	}
	err := s.repo.SaveChat(chat)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}

	// Add creator as a user and admin
	userID, err := uuid.Parse(creatorID)
	if err != nil {
		slog.Error("Invalid creator ID", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddUserToChat(chat.ID, userID)
	if err != nil {
		slog.Error("Failed to add user to chat", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}
	err = s.repo.AddAdmin(chat.ID, userID)
	if err != nil {
		slog.Error("Failed to add admin to chat", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}

	slog.Info("Chat created successfully", "chatID", chat.ID)
	return chat.ID.String(), nil
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
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Chat deleted successfully", "chatID", chatID)
	return nil
}

func (s *ChatManagementService) UpdateChat(ctx context.Context, chatID, name, description, userID string) error {
	slog.Info("UpdateChat called", "chatID", chatID, "name", name, "description", description, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, chatID, userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	id, err := uuid.Parse(chatID)
	if err != nil {
		slog.Error("Invalid chat ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.UpdateChat(id, name, description)
	if err != nil {
		slog.Error("Failed to update chat", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Chat updated successfully", "chatID", chatID)
	return nil
}

func (s *ChatManagementService) JoinChat(ctx context.Context, chatID, userID string) error {
	slog.Info("JoinChat called", "chatID", chatID, "userID", userID)
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
	err = s.repo.AddUserToChat(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to join chat", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("User joined chat successfully", "chatID", chatID, "userID", userID)
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
	err = s.repo.RemoveUserFromChat(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to kick user from chat", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
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
	isMember, err := s.repo.IsMember(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to check member status", "error", err.Error())
		return false, status.Error(codes.Internal, err.Error())
	}
	return isMember, nil
}

func (s *ChatManagementService) MakeAdmin(ctx context.Context, chatID, userID, requestingUserID string) error {
	slog.Info("MakeAdmin called", "chatID", chatID, "userID", userID, "requestingUserID", requestingUserID)
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
	err = s.repo.AddAdmin(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Admin added successfully", "chatID", chatID, "userID", userID)
	return nil
}

func (s *ChatManagementService) DeleteAdmin(ctx context.Context, chatID, userID, requestingUserID string) error {
	slog.Info("DeleteAdmin called", "chatID", chatID, "userID", userID, "requestingUserID", requestingUserID)
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
	err = s.repo.RemoveAdmin(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Admin removed successfully", "chatID", chatID, "userID", userID)
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
	isAdmin, err := s.repo.IsAdmin(chatUUID, userUUID)
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.Internal, err.Error())
	}
	return isAdmin, nil
}

func (s *ChatManagementService) isAdmin(ctx context.Context, chatID, userID string) (bool, error) {
	return s.IsAdmin(ctx, chatID, userID)
}
