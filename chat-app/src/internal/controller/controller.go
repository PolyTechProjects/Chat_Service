package controller

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"example.com/chat-app/src/internal/client"
	"example.com/chat-app/src/internal/dto"
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

func (ws *WebsocketController) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occurred while trying to connect to websocket server.")
		return
	}

	token := r.Header.Get("Authorization")
	_, err = ws.authClient.PerformAuthorize(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	extractResp, err := ws.authClient.PerformUserIdExtraction(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ws.messageService.ReadMessages(userId, wsConnection, ws.broadcastChannel, token)
	wsConnection.Close()
}

func (ws *WebsocketController) StartListeningFileChannel() {
	ws.messageService.ListenFileChannel()
}

func (ws *WebsocketController) StartBroadcasting() {
	subscriber := ws.messageService.SubscribeToMessageChannel()
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Not Available message-channel")
		return
	}
	slog.Info("Available message-channel")
	for {
		messageWithToken, err := receiveMessageFromRedis(subscriber)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Info("Received message")

		message := messageWithToken.Message
		token := messageWithToken.Token

		channelID := message.ChatRoomId
		channelUsers, err := ws.channelMgmtClient.PerformGetChanUsers(channelID, token)
		if err == nil && len(channelUsers) > 0 {
			messageModel, err := models.MapRequestToMessage(&message)
			if err != nil {
				slog.Error("Error has occurred while mapping message to messageModel", err)
				continue
			}
			ws.messageService.Broadcast(channelUsers, messageModel)
			continue
		}

		chatUsers, err := ws.chatMgmtClient.PerformGetChatUsers(channelID, token)
		if err == nil && len(chatUsers) > 0 {
			messageModel, err := models.MapRequestToMessage(&message)
			if err != nil {
				slog.Error("Error has occurred while mapping message to messageModel", err)
				continue
			}
			ws.messageService.Broadcast(chatUsers, messageModel)
			continue
		}

		slog.Error("No users found in either channel or chat management for broadcasting")
	}
}

func receiveMessageFromRedis(subscriber *redis.PubSub) (*dto.MessageWithToken, error) {
	messageWithToken := &dto.MessageWithToken{}
	slog.Info("Waiting for message")
	channel := subscriber.Channel()
	receivedMessage := <-channel
	slog.Info(receivedMessage.Payload)
	err := json.Unmarshal([]byte(receivedMessage.Payload), messageWithToken)
	if err != nil {
		return nil, err
	}
	return messageWithToken, nil
}
