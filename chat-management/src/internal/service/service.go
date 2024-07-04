package service

import (
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

func (s *ChatManagementService) GetChat(req *dto.GetChatRequest) (*dto.GetChatResponse, error) {
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

func (s *ChatManagementService) GetChatWithAdmins(req *dto.GetChatRequest) (*dto.GetChatResponse, error) {
	getResp, err := s.GetChat(req)
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

func (s *ChatManagementService) CreateChat(req *dto.CreateChatRequest) (dto.ChatResponse, error) {
	slog.Info("CreateChat called", "name", req.Name, "description", req.Description, "creatorID", req.CreatorId)
	chat := models.Chat{
		Id:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatorId:   req.CreatorId,
	}
	err := s.repo.SaveChat(&chat)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	userChat := models.UserChat{
		ChatId: chat.Id,
		UserId: req.CreatorId,
	}
	err = s.repo.AddUserToChat(&userChat)
	if err != nil {
		slog.Error("Failed to add user to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}
	admin := models.Admin{
		ChatId: chat.Id,
		UserId: req.CreatorId,
	}
	err = s.repo.AddAdmin(&admin)
	if err != nil {
		slog.Error("Failed to add admin to chat", "error", err.Error())
		return dto.ChatResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	slog.Info("Chat created successfully", "chatID", chat.Id)
	return dto.ChatResponse{ChatId: chat.Id.String()}, nil
}

func (s *ChatManagementService) DeleteChat(req *dto.DeleteChatRequest) error {
	slog.Info("DeleteChat called", "chatID", req.ChatId, "userID", req.UserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}

	err := s.repo.DeleteChat(req.ChatId)
	if err != nil {
		slog.Error("Failed to delete chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Chat deleted successfully", "chatID", req.ChatId)
	return nil
}

func (s *ChatManagementService) UpdateChat(req *dto.UpdateChatRequest) error {
	slog.Info("UpdateChat called", "chatID", req.ChatId, "name", req.Name, "description", req.Description, "userID", req.UserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
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

func (s *ChatManagementService) JoinChat(req *dto.JoinChatRequest) error {
	slog.Info("JoinChat called", "chatID", req.ChatId, "userID", req.UserId)
	userChat := models.UserChat{
		ChatId: req.ChatId,
		UserId: req.UserId,
	}
	err := s.repo.AddUserToChat(&userChat)
	if err != nil {
		slog.Error("Failed to add user to chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User joined chat successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) LeaveChat(req *dto.LeaveChatRequest) error {
	slog.Info("LeaveChat called", "chatID", req.ChatId, "userID", req.UserId)
	userChat := models.UserChat{
		ChatId: req.ChatId,
		UserId: req.UserId,
	}
	err := s.repo.RemoveUserFromChat(&userChat)
	if err != nil {
		slog.Error("Failed to remove user from chat", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User left chat successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) InviteUser(req *dto.InviteUserRequest) error {
	slog.Info("InviteUser called", "chatID", req.ChatId, "userID", req.UserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	userChat := models.UserChat{
		ChatId: req.ChatId,
		UserId: req.RequestingUserId,
	}
	err := s.repo.AddUserToChat(&userChat)
	if err != nil {
		slog.Error("Failed to invite user", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User invited successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) KickUser(req *dto.KickUserRequest) error {
	slog.Info("KickUser called", "chatID", req.ChatId, "userID", req.UserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	userChat := models.UserChat{
		ChatId: req.ChatId,
		UserId: req.RequestingUserId,
	}
	err := s.repo.RemoveUserFromChat(&userChat)
	if err != nil {
		slog.Error("Failed to kick user", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("User kicked successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) MakeAdmin(req *dto.AdminRequest) error {
	slog.Info("MakeAdmin called", "chatID", req.ChatId, "userID", req.UserId, "requestingUserId", req.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	admin := models.Admin{
		ChatId: req.ChatId,
		UserId: req.RequestingUserId,
	}
	err := s.repo.AddAdmin(&admin)
	if err != nil {
		slog.Error("Failed to add admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin added successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) DeleteAdmin(req *dto.AdminRequest) error {
	slog.Info("DeleteAdmin called", "chatID", req.ChatId, "userID", req.UserId, "requestingUserId", req.RequestingUserId)
	isAdminReq := &dto.IsAdminRequest{ChatId: req.ChatId, UserId: req.UserId}
	if isAdmin, err := s.IsAdmin(isAdminReq); err != nil || !isAdmin {
		slog.Error("Permission denied or error checking admin status", "error", err)
		return status.Error(codes.PermissionDenied, "permission denied or error checking admin status")
	}
	admin := models.Admin{
		ChatId: req.ChatId,
		UserId: req.RequestingUserId,
	}
	err := s.repo.RemoveAdmin(&admin)
	if err != nil {
		slog.Error("Failed to remove admin", "error", err.Error())
		return status.Error(codes.InvalidArgument, err.Error())
	}
	slog.Info("Admin removed successfully", "chatID", req.ChatId, "userID", req.UserId)
	return nil
}

func (s *ChatManagementService) IsAdmin(adminReq *dto.IsAdminRequest) (bool, error) {
	slog.Info("IsAdmin called", "chatID", adminReq.ChatId, "userID", adminReq.UserId)
	admin := models.Admin{
		ChatId: adminReq.ChatId,
		UserId: adminReq.UserId,
	}
	isAdmin, err := s.repo.IsAdmin(&admin)
	if err != nil {
		slog.Error("Failed to check admin status", "error", err.Error())
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return isAdmin, nil
}
