package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/channel-management/src/internal/client"
	"example.com/channel-management/src/internal/dto"
	"example.com/channel-management/src/internal/service"
)

type ChannelManagementController struct {
	service    *service.ChannelManagementService
	authClient *client.AuthGRPCClient
}

func NewChannelManagementController(service *service.ChannelManagementService, authClient *client.AuthGRPCClient) *ChannelManagementController {
	return &ChannelManagementController{
		service:    service,
		authClient: authClient,
	}
}

func (c *ChannelManagementController) CreateChannelHandler(w http.ResponseWriter, r *http.Request) {
	var channelReq dto.CreateChannelRequest
	err := json.NewDecoder(r.Body).Decode(&channelReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, channelReq.CreatorId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	channel, err := c.service.CreateChannel(&channelReq)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	channelResp, err := json.Marshal(channel)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(channelResp)
}

func (c *ChannelManagementController) DeleteChannelHandler(w http.ResponseWriter, r *http.Request) {
	var channelReq dto.DeleteChannelRequest
	err := json.NewDecoder(r.Body).Decode(&channelReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, channelReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.DeleteChannel(&channelReq)
	if err != nil {
		slog.Error("Failed to delete chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) UpdateChannelHandler(w http.ResponseWriter, r *http.Request) {
	var channelReq dto.UpdateChannelRequest
	err := json.NewDecoder(r.Body).Decode(&channelReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, channelReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.UpdateChannel(&channelReq)
	if err != nil {
		slog.Error("Failed to update chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) JoinChannelHandler(w http.ResponseWriter, r *http.Request) {
	var joinReq dto.JoinChannelRequest
	err := json.NewDecoder(r.Body).Decode(&joinReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, joinReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.JoinChannel(&joinReq)
	if err != nil {
		slog.Error("Failed to join chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) LeaveChannelHandler(w http.ResponseWriter, r *http.Request) {
	var leaveReq dto.LeaveChannelRequest
	err := json.NewDecoder(r.Body).Decode(&leaveReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, leaveReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.LeaveChannel(&leaveReq)
	if err != nil {
		slog.Error("Failed to leave chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) InviteUserHandler(w http.ResponseWriter, r *http.Request) {
	var inviteReq dto.InviteUserRequest
	err := json.NewDecoder(r.Body).Decode(&inviteReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, inviteReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.InviteUser(&inviteReq)
	if err != nil {
		slog.Error("Failed to invite user", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) KickUserHandler(w http.ResponseWriter, r *http.Request) {
	var kickReq dto.KickUserRequest
	err := json.NewDecoder(r.Body).Decode(&kickReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, kickReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.KickUser(&kickReq)
	if err != nil {
		slog.Error("Failed to kick user", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
	var adminReq dto.AdminRequest
	err := json.NewDecoder(r.Body).Decode(&adminReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, adminReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	isAdminReq := dto.IsAdminRequest{
		ChannelId: adminReq.ChannelId,
		UserId:    adminReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", adminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
		return
	}

	err = c.service.MakeAdmin(&adminReq)
	if err != nil {
		slog.Error("Failed to make admin", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) DeleteAdminHandler(w http.ResponseWriter, r *http.Request) {
	var adminReq dto.AdminRequest
	err := json.NewDecoder(r.Body).Decode(&adminReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, adminReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	isAdminReq := dto.IsAdminRequest{
		ChannelId: adminReq.ChannelId,
		UserId:    adminReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", adminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
		return
	}

	err = c.service.DeleteAdmin(&adminReq)
	if err != nil {
		slog.Error("Failed to delete admin", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChannelManagementController) GetChannelHandler(w http.ResponseWriter, r *http.Request) {
	var channelReq dto.GetChannelRequest
	err := json.NewDecoder(r.Body).Decode(&channelReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r.Context(), r, channelReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	channel, err := c.service.GetChannel(&channelReq)
	if err != nil {
		slog.Error("Failed to get channel", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	channelResp, err := json.Marshal(channel)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(channelResp)
}
