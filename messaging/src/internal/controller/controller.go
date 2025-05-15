package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"example.com/messaging/src/internal/client"
	"example.com/messaging/src/internal/dto"
	"example.com/messaging/src/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageHistoryController struct {
	messageHistoryService *service.MessageHistoryService
	authClient            *client.AuthGRPCClient
}

func NewMessageHistoryController(messageHistoryService *service.MessageHistoryService, authClient *client.AuthGRPCClient) *MessageHistoryController {
	return &MessageHistoryController{
		messageHistoryService: messageHistoryService,
		authClient:            authClient,
	}
}

func (m *MessageHistoryController) GetDirectHistoryHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	destinationId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = m.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := m.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages, err := m.messageHistoryService.GetDirectHistory(destinationId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response, err := json.Marshal(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(response)
}

func (m *MessageHistoryController) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	destinationId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = m.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := m.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages, err := m.messageHistoryService.GetHistory(destinationId, userId)
	if errors.Is(err, fmt.Errorf("permission denied")) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response, err := json.Marshal(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(response)
}

func (c *MessageHistoryController) DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	messageId, err := uuid.Parse(segments[len(segments)-1])
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

	err = c.messageHistoryService.DeleteMessage(messageId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func (c *MessageHistoryController) EditMessageHandler(w http.ResponseWriter, r *http.Request) {
	req := &dto.EditMessageRequest{}
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

	err = c.messageHistoryService.EditMessage(req.MessageId, req.Body, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

type WebsocketController struct {
	messageService *service.MessageService
	authClient     *client.AuthGRPCClient
	chatClient     *client.ChatGRPCClient
}

func NewWebsocketController(messageService *service.MessageService, authClient *client.AuthGRPCClient, chatClient *client.ChatGRPCClient) *WebsocketController {
	return &WebsocketController{
		messageService: messageService,
		authClient:     authClient,
		chatClient:     chatClient,
	}
}

func (ws *WebsocketController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occurred while trying to connect to websocket server: " + err.Error())
		return
	}
	slog.Debug("Connected to websocket server")
	defer wsConnection.Close()

	path := r.URL.Path
	segments := strings.Split(path, "/")
	chatId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("token") {
		wsConnection.Close()
		http.Error(w, "URL query params are invalid", http.StatusBadRequest)
	}
	accessToken := params.Get("token")

	_, err = ws.authClient.PerformAuthorizeDirect(accessToken)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := ws.authClient.PerformExtractUserIdDirect(accessToken)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	verifyPersistanceResp, err := ws.chatClient.PerformVerifyUserPersistance(chatId.String(), userId.String())
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !verifyPersistanceResp.IsVerified {
		wsConnection.Close()
		http.Error(w, fmt.Errorf("permission denied").Error(), http.StatusForbidden)
		return
	}

	ws.messageService.ReadMessages(wsConnection, userId, chatId)
}

func (ws *WebsocketController) SendDirectMessageHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occurred while trying to connect to websocket server: " + err.Error())
		return
	}
	slog.Debug("Connected to websocket server")
	defer wsConnection.Close()

	path := r.URL.Path
	segments := strings.Split(path, "/")
	destinationId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("token") {
		wsConnection.Close()
		http.Error(w, "URL query params are invalid", http.StatusBadRequest)
	}
	accessToken := params.Get("token")

	_, err = ws.authClient.PerformAuthorizeDirect(accessToken)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := ws.authClient.PerformExtractUserIdDirect(accessToken)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = ws.messageService.ReadMessages(wsConnection, userId, destinationId)
	if err != nil {
		wsConnection.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
