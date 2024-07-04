package controller

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/chat-app/src/internal/client"
	"example.com/chat-app/src/internal/dto"
	"example.com/chat-app/src/internal/errors"
	"example.com/chat-app/src/internal/models"
	"example.com/chat-app/src/internal/service"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageHistoryController struct {
	messageHistoryService *service.MessageHistoryService
}

func NewMessageHistoryController(messageHistoryService *service.MessageHistoryService) *MessageHistoryController {
	return &MessageHistoryController{
		messageHistoryService: messageHistoryService,
	}
}

func (m *MessageHistoryController) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	chatRoomId, err := uuid.Parse(strings.Split(r.URL.Path, "/")[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	messages, err := m.messageHistoryService.GetHistory(chatRoomId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var messageResps []dto.MessageResponse

	for _, message := range messages {
		messageResps = append(messageResps, *models.MapMessageToResponse(&message))
	}

	response, err := json.Marshal(messageResps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(response)
}

type WebsocketController struct {
	messageService    *service.MessageService
	broadcastChannel  chan *models.Message
	authClient        *client.AuthGRPCClient
	channelMgmtClient *client.ChanMgmtGRPCClient
	chatMgmtClient    *client.ChatMgmtGRPCClient
}

func NewWebsocketController(messageService *service.MessageService, authClient *client.AuthGRPCClient, channelMgmtClient *client.ChanMgmtGRPCClient, chatMgmtClient *client.ChatMgmtGRPCClient) *WebsocketController {
	return &WebsocketController{
		messageService:    messageService,
		broadcastChannel:  make(chan *models.Message),
		authClient:        authClient,
		channelMgmtClient: channelMgmtClient,
		chatMgmtClient:    chatMgmtClient,
	}
}

func (ws *WebsocketController) extractUserIdAndTokens(wsConnection *websocket.Conn, r *http.Request) (uuid.UUID, string, string) {
	header := r.Header.Get("Authorization")
	accessToken := strings.Split(header, "Bearer ")[1]

	cookie, err := r.Cookie("X-Refresh-Token")
	if err != nil {
		slog.Error("X-Refresh-Token cookie not found")
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: "X-Refresh-Token cookie not found"})
		return uuid.Nil, "", ""
	}
	refreshToken := cookie.Value

	userIdH := r.Header.Get("X-User-Id")
	if userIdH == "" {
		slog.Error("X-User-Id header not found")
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: "X-User-Id header not found"})
		return uuid.Nil, "", ""
	}

	authorizeResp, err := ws.authClient.PerformAuthorize(r.Context(), accessToken, refreshToken, userIdH)
	if err != nil {
		slog.Error(err.Error())
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
		return uuid.Nil, "", ""
	}
	slog.Debug("Authorized")

	slog.Debug(fmt.Sprintf("Extracted user id: %s", authorizeResp.UserId))
	userId, err := uuid.Parse(authorizeResp.UserId)
	if err != nil {
		slog.Error(err.Error())
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
		return uuid.Nil, "", ""
	}
	return userId, accessToken, refreshToken
}

func (ws *WebsocketController) SendMessageInChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occurred while trying to connect to websocket server.")
		return
	}
	slog.Debug("Connected to websocket server")
	defer wsConnection.Close()
	userId, accessToken, refreshToken := ws.extractUserIdAndTokens(wsConnection, r)

	err = ws.messageService.ReadMessagesFromChatRoom(userId, wsConnection, accessToken, refreshToken)
	if err != nil {
		if e.Is(err, errors.ErrDatabaseInternalError) {
			slog.Debug(fmt.Sprintf("%v: %v", http.StatusInternalServerError, err.Error()))
			wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
			return
		}
		slog.Debug(fmt.Sprintf("%v: %v", http.StatusBadRequest, err.Error()))
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
		return
	}
}

func (ws *WebsocketController) SendMessageInChannelHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occurred while trying to connect to websocket server.")
		return
	}
	slog.Debug("Connected to websocket server")
	defer wsConnection.Close()
	userId, accessToken, refreshToken := ws.extractUserIdAndTokens(wsConnection, r)

	err = ws.messageService.ReadMessagesFromChannel(userId, wsConnection, accessToken, refreshToken)
	if err != nil {
		if e.Is(err, errors.ErrDatabaseInternalError) {
			slog.Debug(fmt.Sprintf("%v: %v", http.StatusInternalServerError, err.Error()))
			wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
			return
		}
		slog.Debug(fmt.Sprintf("%v: %v", http.StatusBadRequest, err.Error()))
		wsConnection.WriteJSON(models.ErrorMessageResponse{Error: err.Error()})
		return
	}
}

func (ws *WebsocketController) StartListeningFileChannel() {
	ws.messageService.ListenFileChannel()
}

func (ws *WebsocketController) StartBroadcastingToChatRooms() {
	subscriber := ws.messageService.SubscribeToMessageChannel(ws.messageService.RedisChannelForChatRoomMessagesName)
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Channel Not Available", "channel", ws.messageService.RedisChannelForChatRoomMessagesName)
		return
	}
	slog.Info("Channel Available", "channel", ws.messageService.RedisChannelForChatRoomMessagesName)
	for {
		messageWithTokens, err := receiveMessageFromRedis(subscriber)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Info("Received message")

		message := messageWithTokens.Message
		slog.Debug(fmt.Sprintf("Message: %v", message))
		accessToken := messageWithTokens.AccessToken
		refreshToken := messageWithTokens.RefreshToken

		chatRoomId := message.ChatRoomId.String()
		chatUsers, err := ws.chatMgmtClient.PerformGetChatUsers(chatRoomId, accessToken, refreshToken, message.SenderId.String())
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		readyMessage := models.ReadyMessage{
			Message:      message,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ReceiversIds: chatUsers,
		}
		ws.messageService.Broadcast(chatUsers, &readyMessage)
	}
}

func (ws *WebsocketController) StartBroadcastingToChannels() {
	subscriber := ws.messageService.SubscribeToMessageChannel(ws.messageService.RedisChannelForChannelMessagesName)
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Channel Not Available", "channel", ws.messageService.RedisChannelForChannelMessagesName)
		return
	}
	slog.Info("Channel Available", "channel", ws.messageService.RedisChannelForChannelMessagesName)
	for {
		messageWithTokens, err := receiveMessageFromRedis(subscriber)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Info("Received message")

		message := messageWithTokens.Message
		slog.Debug(fmt.Sprintf("Message: %v", message))
		accessToken := messageWithTokens.AccessToken
		refreshToken := messageWithTokens.RefreshToken

		channelId := message.ChatRoomId.String()
		isAdmin, err := ws.channelMgmtClient.PerformIsAdmin(channelId, accessToken, refreshToken, message.SenderId.String())
		if err != nil || !isAdmin {
			slog.Error(err.Error())
			continue
		}
		channelUsers, err := ws.channelMgmtClient.PerformGetChanUsers(channelId, accessToken, refreshToken, message.SenderId.String())
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		readyMessage := models.ReadyMessage{
			Message:      message,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ReceiversIds: channelUsers,
		}
		ws.messageService.Broadcast(channelUsers, &readyMessage)
	}
}

func receiveMessageFromRedis(subscriber *redis.PubSub) (*models.MessageWithTokens, error) {
	messageWithTokens := &models.MessageWithTokens{}
	slog.Debug("Waiting for message")
	channel := subscriber.Channel()
	receivedMessage := <-channel
	slog.Debug(receivedMessage.Payload)
	err := json.Unmarshal([]byte(receivedMessage.Payload), messageWithTokens)
	if err != nil {
		return nil, err
	}
	return messageWithTokens, nil
}
