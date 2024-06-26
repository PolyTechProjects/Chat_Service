package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/channel-management/src/internal/dto"
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

func (s *ChannelManagementService) GetChannel(ctx context.Context, req *dto.GetChannelRequest) (*dto.GetChannelResponse, error) {
	slog.Info("GetChannel called", "channelID", req.ChannelId)
	channel, err := s.repo.FindById(req.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	users, err := s.repo.GetChanUsers(req.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getResp := &dto.GetChannelResponse{
		Channel: channel,
		Users:   users,
	}
	return getResp, nil
}

func (s *ChannelManagementService) GetChannelWithAdmins(ctx context.Context, req *dto.GetChannelRequest) (*dto.GetChannelResponse, error) {
	getResp, err := s.GetChannel(ctx, req)
	if err != nil {
		return nil, err
	}
	admins, err := s.repo.GetChannelAdmins(req.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getResp.Admins = admins
	return getResp, nil
}

func (s *ChannelManagementService) CreateChannel(ctx context.Context, name, description, creatorID string) (string, error) {
	slog.Info("CreateChannel called", "name", name, "description", description, "creatorID", creatorID)
	creatorUUID, err := uuid.Parse(creatorID)
	if err != nil {
		slog.Error("Invalid creator ID", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	channelDTO := dto.CreateChannelDTO{
		Name:        name,
		Description: description,
		CreatorId:   creatorUUID,
	}
	channelID, err := s.repo.SaveChannel(channelDTO)
	if err != nil {
		slog.Error("Failed to create channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	// Add creator as a user and admin
	err = s.repo.AddUserToChannel(dto.AddUserDTO{ChannelId: channelID, UserId: creatorUUID})
	if err != nil {
		slog.Error("Failed to add user to channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.AddAdmin(dto.AdminDTO{ChannelId: channelID, UserId: creatorUUID})
	if err != nil {
		slog.Error("Failed to add admin to channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("Channel created successfully", "channelID", channelID)
	return channelID.String(), nil
}

func (s *ChannelManagementService) DeleteChannel(ctx context.Context, channelID, userID string) error {
	slog.Info("DeleteChannel called", "channelID", channelID, "userID", userID)
	if isAdmin, err := s.isAdmin(ctx, channelID, userID); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	err = s.repo.DeleteChannel(channelUUID)
	if err != nil {
		slog.Error("Failed to delete channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel ID", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	updateDTO := dto.UpdateChannelDTO{
		Id:          channelUUID,
		Name:        name,
		Description: description,
	}
	err = s.repo.UpdateChannel(updateDTO)
	if err != nil {
		slog.Error("Failed to update channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
	addUserDTO := dto.AddUserDTO{
		ChannelId: channelUUID,
		UserId:    userUUID,
	}
	err = s.repo.AddUserToChannel(addUserDTO)
	if err != nil {
		slog.Error("Failed to join channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
	removeUserDTO := dto.RemoveUserDTO{
		ChannelId: channelUUID,
		UserId:    userUUID,
	}
	err = s.repo.RemoveUserFromChannel(removeUserDTO)
	if err != nil {
		slog.Error("Failed to kick user from channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
		return false, status.Error(codes.InvalidArgument, err.Error())
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
	addAdminDTO := dto.AdminDTO{
		ChannelId: channelUUID,
		UserId:    userUUID,
	}
	err = s.repo.AddAdmin(addAdminDTO)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
	removeAdminDTO := dto.AdminDTO{
		ChannelId: channelUUID,
		UserId:    userUUID,
	}
	err = s.repo.RemoveAdmin(removeAdminDTO)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
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
		return false, status.Error(codes.InvalidArgument, err.Error())
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
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return userIDs, nil
}
