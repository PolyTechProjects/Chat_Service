package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"example.com/chat/src/internal/client"
	"example.com/chat/src/internal/dto"
	"example.com/chat/src/internal/service"
	"example.com/chat/src/internal/validator"
	"github.com/google/uuid"
)

type ChatController struct {
	chatService *service.ChatService
	authClient  *client.AuthGRPCClient
}

func NewChatController(chatService *service.ChatService, authClient *client.AuthGRPCClient) *ChatController {
	return &ChatController{
		chatService: chatService,
		authClient:  authClient,
	}
}

func (c *ChatController) GetChatHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	chatId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := c.chatService.GetChat(chatId, userId)
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
	w.Write(chatResp)
}

func (c *ChatController) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	chatId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.DeleteChat(chatId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *ChatController) CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateChatRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if len(req.ParticipantsIds) < 2 {
		http.Error(w, "Not enough participants", http.StatusBadRequest)
		return
	}

	err = validator.ValidateChatName(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := c.chatService.CreateChat(req)
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
	w.Write(chatResp)
}

func (c *ChatController) EditChatHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.EditChatRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validator.ValidateChatName(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.EditChat(req, userId)
	if errors.Is(err, fmt.Errorf("permission denied")) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := c.chatService.GetChat(req.ChatId, userId)
	if errors.Is(err, fmt.Errorf("permission denied")) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
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
	w.Write(chatResp)
}

func (c *ChatController) JoinChatHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	joinLink := segments[len(segments)-1]

	_, err := c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := c.chatService.JoinChat(joinLink, userId)
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
	w.Write(chatResp)
}

func (c *ChatController) AddUsersInChatHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.AddUsersRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.AddUsers(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *ChatController) DeleteUsersInChatHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.DeleteUsersRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.DeleteUsers(req.ChatId, req.UserIds, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (c *ChatController) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateRoleRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := c.chatService.CreateRole(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	roleResp, err := json.Marshal(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(roleResp)
}

func (c *ChatController) EditRoleHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.UpdateRoleRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := c.chatService.EditRole(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	roleResp, err := json.Marshal(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(roleResp)
}

func (c *ChatController) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	chatId, err := uuid.Parse(segments[len(segments)-3])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	roleId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.DeleteRole(chatId, userId, roleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChatController) SetRoleHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.SetRoleRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.SetRole(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChatController) ChangeUserNicknameHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.ChangeUserNicknameRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.ChangeUserNickname(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *ChatController) GetDirectChatHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	firstUserId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	secondUserId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := c.chatService.GetDirectChat(firstUserId, secondUserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(chat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (c *ChatController) CreateDirectChatHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.CreateDirectChatRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	directChat, err := c.chatService.CreateDirectChat(req.UserId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	directChatResp, err := json.Marshal(directChat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(directChatResp)
}

func (c *ChatController) LeaveFromChatsHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.LeaveFromChatsRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.chatService.LeaveFromChats(req, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (c *ChatController) GetAllAvailableChatsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chats, err := c.chatService.GetChats(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chatsResp, err := json.Marshal(chats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(chatsResp)
}
