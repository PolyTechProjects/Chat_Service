package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/channel-management/src/internal/models"
	"example.com/channel-management/src/internal/repository"
	"github.com/google/uuid"
)

type ChannelManagementService struct {
	repo repository.ChannelRepository
}

func New(repo repository.ChannelRepository) *ChannelManagementService {
	return &ChannelManagementService{
		repo: repo,
	}
}

func (s *ChannelManagementService) CreateChannel(ctx context.Context, name, description, creatorID string) (string, error) {
	slog.Info("CreateChannel called", "name", name, "description", description, "creatorID", creatorID)
	channel := models.Channel{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
	}
	err := s.repo.SaveChannel(channel)
	if err != nil {
		slog.Error("Failed to create channel", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}

	// Add creator as a user and admin
	userID, err := uuid.Parse(creatorID)
	if err != nil {
		slog.Error("Invalid creator ID", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddUserToChannel(channel.ID, userID)
	if err != nil {
		slog.Error("Failed to add user to channel", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}
	err = s.repo.AddAdmin(channel.ID, userID)
	if err != nil {
		slog.Error("Failed to add admin to channel", "error", err.Error())
		return "", status.Error(codes.Internal, err.Error())
	}

	slog.Info("Channel created successfully", "channelID", channel.ID)
	return channel.ID.String(), nil
}

func (s *ChannelManagementService) DeleteChannel(ctx context.Context, channelID, userID string) error {
	slog.Info("DeleteChannel called", "channelID", channelID, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, channelID, userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	id, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.DeleteChannel(id)
	if err != nil {
		slog.Error("Failed to delete channel", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Channel deleted successfully", "channelID", channelID)
	return nil
}

func (s *ChannelManagementService) UpdateChannel(ctx context.Context, channelID, name, description, userID string) error {
	slog.Info("UpdateChannel called", "channelID", channelID, "name", name, "description", description, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, channelID, userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	id, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.UpdateChannel(id, name, description)
	if err != nil {
		slog.Error("Failed to update channel", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Channel updated successfully", "channelID", channelID)
	return nil
}

func (s *ChannelManagementService) JoinChannel(ctx context.Context, channelID, userID string) error {
	slog.Info("JoinChannel called", "channelID", channelID, "userID", userID)
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddUserToChannel(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to join channel", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("User joined channel successfully", "channelID", channelID, "userID", userID)
	return nil
}

func (s *ChannelManagementService) KickUser(ctx context.Context, channelID, userID, requestingUserID string) error {
	slog.Info("KickUser called", "channelID", channelID, "userID", userID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, channelID, requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.RemoveUserFromChannel(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to kick user from channel", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("User kicked from channel successfully", "channelID", channelID, "userID", userID)
	return nil
}

func (s *ChannelManagementService) CanWrite(ctx context.Context, channelID, userID string) (bool, error) {
	slog.Info("CanWrite called", "channelID", channelID, "userID", userID)
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	isAdmin, err := s.repo.IsAdmin(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.Internal, err.Error())
	}
	return isAdmin, nil
}

func (s *ChannelManagementService) MakeAdmin(ctx context.Context, channelID, userID, requestingUserID string) error {
	slog.Info("MakeAdmin called", "channelID", channelID, "userID", userID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, channelID, requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddAdmin(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Admin added successfully", "channelID", channelID, "userID", userID)
	return nil
}

func (s *ChannelManagementService) DeleteAdmin(ctx context.Context, channelID, userID, requestingUserID string) error {
	slog.Info("DeleteAdmin called", "channelID", channelID, "userID", userID, "requestingUserID", requestingUserID)
	if isAdmin, err := s.isAdmin(ctx, channelID, requestingUserID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.RemoveAdmin(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.Internal, err.Error())
	}
	slog.Info("Admin removed successfully", "channelID", channelID, "userID", userID)
	return nil
}

func (s *ChannelManagementService) IsAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	slog.Info("IsAdmin called", "channelID", channelID, "userID", userID)
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Invalid user ID", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	isAdmin, err := s.repo.IsAdmin(channelUUID, userUUID)
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.Internal, err.Error())
	}
	return isAdmin, nil
}

func (s *ChannelManagementService) isAdmin(ctx context.Context, channelID, userID string) (bool, error) {
	return s.IsAdmin(ctx, channelID, userID)
}

func (s *ChannelManagementService) GetChanUsers(ctx context.Context, chanID string) ([]string, error) {
	slog.Info("GetChanUsers called", "channelID", chanID)
	chanUUID, err := uuid.Parse(chanID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userIDs, err := s.repo.GetChanUsers(chanUUID)
	if err != nil {
		slog.Error("Failed to get channel users", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return userIDs, nil
}
