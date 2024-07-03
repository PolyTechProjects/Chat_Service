package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/chat-management/src/internal/client"
	"example.com/chat-management/src/internal/dto"
	"example.com/chat-management/src/internal/service"
)

type ChatManagementController struct {
	service    *service.ChatManagementService
	authClient *client.AuthGRPCClient
}

func NewChatManagementController(service *service.ChatManagementService, authClient *client.AuthGRPCClient) *ChatManagementController {
	return &ChatManagementController{
		service:    service,
		authClient: authClient,
	}
}

func (c *ChatManagementController) CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq dto.CreateChatRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		slog.Error("Failed to decode request", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, chatReq.CreatorId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	chat, err := c.service.CreateChat(&chatReq)
	if err != nil {
		slog.Error("Failed to create chat", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	chatResp, err := json.Marshal(chat)
	if err != nil {
		slog.Error("Failed to marshal response", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
	w.Write(chatResp)
}

func (c *ChatManagementController) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq dto.DeleteChatRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, chatReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.DeleteChat(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) GetChatHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq dto.GetChatRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, chatReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	chat, err := c.service.GetChat(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	chatResp, err := json.Marshal(chat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
	w.Write(chatResp)
}

func (c *ChatManagementController) UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
	var chatReq dto.UpdateChatRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, chatReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.UpdateChat(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) JoinChatHandler(w http.ResponseWriter, r *http.Request) {
	var joinReq dto.JoinChatRequest
	err := json.NewDecoder(r.Body).Decode(&joinReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, joinReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.JoinChat(&joinReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) LeaveChatHandler(w http.ResponseWriter, r *http.Request) {
	var leaveReq dto.LeaveChatRequest
	err := json.NewDecoder(r.Body).Decode(&leaveReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, leaveReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = c.service.LeaveChat(&leaveReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) InviteUserHandler(w http.ResponseWriter, r *http.Request) {
	var inviteReq dto.InviteUserRequest
	err := json.NewDecoder(r.Body).Decode(&inviteReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, inviteReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdminReq := dto.IsAdminRequest{
		ChatId: inviteReq.ChatId,
		UserId: inviteReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", isAdminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
	}

	err = c.service.InviteUser(&inviteReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) KickUserHandler(w http.ResponseWriter, r *http.Request) {
	var kickReq dto.KickUserRequest
	err := json.NewDecoder(r.Body).Decode(&kickReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, kickReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdminReq := dto.IsAdminRequest{
		ChatId: kickReq.ChatId,
		UserId: kickReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", isAdminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
	}

	err = c.service.KickUser(&kickReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
	var adminReq dto.AdminRequest
	err := json.NewDecoder(r.Body).Decode(&adminReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, adminReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdminReq := dto.IsAdminRequest{
		ChatId: adminReq.ChatId,
		UserId: adminReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", isAdminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
	}

	err = c.service.MakeAdmin(&adminReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (c *ChatManagementController) DeleteAdminHandler(w http.ResponseWriter, r *http.Request) {
	var adminReq dto.AdminRequest
	err := json.NewDecoder(r.Body).Decode(&adminReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authResp, err := c.authClient.PerformAuthorize(r.Context(), r, adminReq.UserId.String())
	if err != nil {
		slog.Error("Authorization error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	isAdminReq := dto.IsAdminRequest{
		ChatId: adminReq.ChatId,
		UserId: adminReq.UserId,
	}
	if isAdmin, err := c.service.IsAdmin(&isAdminReq); !isAdmin || err != nil {
		slog.Error(fmt.Sprintf("Permission denied: %v not an admin", isAdminReq.UserId))
		http.Error(w, "permission denied", http.StatusForbidden)
	}

	err = c.service.DeleteAdmin(&adminReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}
