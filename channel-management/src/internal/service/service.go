package service

import (
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/channel-management/src/internal/dto"
	"example.com/channel-management/src/internal/models"
	"example.com/channel-management/src/internal/repository"
)

type ChannelManagementService struct {
	repo repository.ChannelRepository
}

func New(repo repository.ChannelRepository) *ChannelManagementService {
	return &ChannelManagementService{
		repo: repo,
	}
}

func (s *ChannelManagementService) GetChannel(getReq *dto.GetChannelRequest) (*dto.GetChannelResponse, error) {
	slog.Info("GetChannel called", "channelID", getReq.ChannelId)
	channel, err := s.repo.FindById(getReq.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	users, err := s.repo.GetChanUsers(getReq.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	admins, err := s.repo.GetChannelAdmins(getReq.ChannelId)
	if err != nil {
		slog.Error("Failed to get channel users", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getResp := &dto.GetChannelResponse{
		Channel: channel,
		Users:   users,
		Admins:  admins,
	}
	return getResp, nil
}

func (s *ChannelManagementService) CreateChannel(channelReq *dto.CreateChannelRequest) (string, error) {
	slog.Info("CreateChannel called", "name", channelReq.Name, "description", channelReq.Description, "creatorID", channelReq.CreatorId)
	channel := models.Channel{
		Id:          uuid.New(),
		Name:        channelReq.Name,
		Description: channelReq.Description,
		CreatorId:   channelReq.CreatorId,
	}
	channelId, err := s.repo.SaveChannel(&channel)
	if err != nil {
		slog.Error("Failed to create channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	userChannel := models.UserChannel{
		ChannelId: channelId,
		UserId:    channelReq.CreatorId,
	}
	err = s.repo.AddUserToChannel(&userChannel)
	if err != nil {
		slog.Error("Failed to add user to channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	admin := models.Admin{
		ChannelId: channelId,
		UserId:    channelReq.CreatorId,
	}
	err = s.repo.AddAdmin(&admin)
	if err != nil {
		slog.Error("Failed to add admin to channel", "error", err.Error())
		return "", status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("Channel created successfully", "channelID", channelId)
	return channelId.String(), nil
}

func (s *ChannelManagementService) DeleteChannel(channelReq *dto.DeleteChannelRequest) error {
	slog.Info("DeleteChannel called", "channelID", channelReq.ChannelId, "userID", channelReq.UserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: channelReq.ChannelId, UserId: channelReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.DeleteChannel(channelReq.ChannelId)
	if err != nil {
		slog.Error("Failed to delete channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Channel deleted successfully", "channelID", channelReq.ChannelId)
	return nil
}

func (s *ChannelManagementService) UpdateChannel(channelReq *dto.UpdateChannelRequest) error {
	slog.Info("UpdateChannel called", "channelID", channelReq.ChannelId, "name", channelReq.Name, "description", channelReq.Description, "userID", channelReq.UserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: channelReq.ChannelId, UserId: channelReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	err := s.repo.UpdateChannel(channelReq)
	if err != nil {
		slog.Error("Failed to update channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Channel updated successfully", "channelID", channelReq.ChannelId)
	return nil
}

func (s *ChannelManagementService) JoinChannel(joinReq *dto.JoinChannelRequest) error {
	slog.Info("JoinChannel called", "channelID", joinReq.ChannelId, "userID", joinReq.UserId)
	userChannel := models.UserChannel{
		ChannelId: joinReq.ChannelId,
		UserId:    joinReq.UserId,
	}
	err := s.repo.AddUserToChannel(&userChannel)
	if err != nil {
		slog.Error("Failed to join channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User joined channel successfully", "channelID", joinReq.ChannelId, "userID", joinReq.UserId)
	return nil
}

func (s *ChannelManagementService) LeaveChannel(leaveReq *dto.LeaveChannelRequest) error {
	slog.Info("LeaveChannel called", "channelID", leaveReq.ChannelId, "userID", leaveReq.UserId)
	userChannel := models.UserChannel{
		ChannelId: leaveReq.ChannelId,
		UserId:    leaveReq.UserId,
	}
	err := s.repo.RemoveUserFromChannel(&userChannel)
	if err != nil {
		slog.Error("Failed to leave channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User left channel successfully", "channelID", leaveReq.ChannelId, "userID", leaveReq.UserId)
	return nil
}

func (s *ChannelManagementService) InviteUser(inviteReq *dto.InviteUserRequest) error {
	slog.Info("InviteUser called", "channelID", inviteReq.ChannelId, "userID", inviteReq.UserId, "requestingUserID", inviteReq.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: inviteReq.ChannelId, UserId: inviteReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	userChannel := models.UserChannel{
		ChannelId: inviteReq.ChannelId,
		UserId:    inviteReq.RequestingUserId,
	}
	err := s.repo.AddUserToChannel(&userChannel)
	if err != nil {
		slog.Error("Failed to invite user to channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User invited to channel successfully", "channelID", inviteReq.ChannelId, "userID", inviteReq.UserId)
	return nil
}

func (s *ChannelManagementService) KickUser(kickReq *dto.KickUserRequest) error {
	slog.Info("KickUser called", "channelID", kickReq.ChannelId, "userID", kickReq.UserId, "requestingUserID", kickReq.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: kickReq.ChannelId, UserId: kickReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	userChannel := models.UserChannel{
		ChannelId: kickReq.ChannelId,
		UserId:    kickReq.RequestingUserId,
	}
	err := s.repo.RemoveUserFromChannel(&userChannel)
	if err != nil {
		slog.Error("Failed to kick user from channel", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User kicked from channel successfully", "channelID", kickReq.ChannelId, "userID", kickReq.UserId)
	return nil
}

func (s *ChannelManagementService) MakeAdmin(adminReq *dto.AdminRequest) error {
	slog.Info("MakeAdmin called", "channelID", adminReq.ChannelId, "userID", adminReq.UserId, "requestingUserID", adminReq.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: adminReq.ChannelId, UserId: adminReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	admin := models.Admin{
		ChannelId: adminReq.ChannelId,
		UserId:    adminReq.RequestingUserId,
	}
	err := s.repo.AddAdmin(&admin)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin added successfully", "channelID", adminReq.ChannelId, "userID", adminReq.UserId)
	return nil
}

func (s *ChannelManagementService) DeleteAdmin(adminReq *dto.AdminRequest) error {
	slog.Info("DeleteAdmin called", "channelID", adminReq.ChannelId, "userID", adminReq.UserId, "requestingUserID", adminReq.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChannelId: adminReq.ChannelId, UserId: adminReq.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	admin := models.Admin{
		ChannelId: adminReq.ChannelId,
		UserId:    adminReq.RequestingUserId,
	}
	err := s.repo.RemoveAdmin(&admin)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin removed successfully", "channelID", adminReq.ChannelId, "userID", adminReq.UserId)
	return nil
}

func (s *ChannelManagementService) IsAdmin(adminReq *dto.IsAdminRequest) (bool, error) {
	slog.Info("IsAdmin called", "channelID", adminReq.ChannelId, "userID", adminReq.UserId)
	admin := models.Admin{
		ChannelId: adminReq.ChannelId,
		UserId:    adminReq.UserId,
	}
	isAdmin, err := s.repo.IsAdmin(&admin)
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return isAdmin, nil
}
