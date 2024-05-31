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
		slog.Error("Error has occured while trying to connect to websocket server.")
	}

	token := r.Header.Get("Authorization")
	_, err = ws.authClient.PerformAuthorize(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	extractResp, err := ws.authClient.PerformUserIdExtraction(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	ws.messageService.ReadMessages(userId, wsConnection, ws.broadcastChannel)
	wsConnection.Close()
}

func (ws *WebsocketController) StartListeningFileChannel() {
	ws.messageService.ListenFileChannel()
}

// func (ws *WebsocketController) StartBroadcasting() {
// 	subscriber := ws.messageService.SubscribeToMessageChannel()
// 	err := subscriber.Ping(context.Background())
// 	if err != nil {
// 		slog.Error("Not Available message-channel")
// 	}
// 	slog.Info("Available message-channel")
// 	for {
// 		//message := <-ws.broadcastChannel
// 		message, err := receiveMessageFromRedis(subscriber)
// 		if err != nil {
// 			slog.Error(err.Error())
// 		}
// 		slog.Info("Received message")
// 		user1 := uuid.MustParse("cfd96643-34a3-466f-9be0-ab079af09419")
// 		user2 := uuid.MustParse("e5b7fa6b-3f2b-45df-bd3d-88f99ab29e40")
// 		user3 := uuid.MustParse("daf3f3e4-98ea-4456-add9-c12906b2f4c0")
// 		//Let's say that message.ChatRoomId is sent to ChatRoomMgmtService which will return List<UserId>
// 		userIds := []uuid.UUID{user1, user2, user3}
// 		ws.messageService.Broadcast(userIds, message)
// 	}
// }

func (ws *WebsocketController) StartBroadcasting() {
	subscriber := ws.messageService.SubscribeToMessageChannel()
	err := subscriber.Ping(context.Background())
	if err != nil {
		slog.Error("Not Available message-channel")
	}
	slog.Info("Available message-channel")
	for {
		message, err := receiveMessageFromRedis(subscriber)
		if err != nil {
			slog.Error(err.Error())
		}
		slog.Info("Received message")

		channelID := message.ChatRoomId.String()
		channelUsers, err := ws.channelMgmtClient.PerformGetChanUsers(channelID)
		if err != nil {
			slog.Error("Error fetching users from channel-management", err)
		}
		chatUsers, err := ws.chatMgmtClient.PerformGetChatUsers(channelID)
		if err != nil {
			slog.Error("Error fetching users from chat-management", err)
		}

		userIDs := mergeUserLists(channelUsers, chatUsers)

		ws.messageService.Broadcast(userIDs, message)
	}
}

func receiveMessageFromRedis(subscriber *redis.PubSub) (*models.Message, error) {
	message := &models.Message{}
	slog.Info("Waiting for message")
	channel := subscriber.Channel()
	receivedMessage := <-channel
	slog.Info(receivedMessage.Payload)
	err := json.Unmarshal([]byte(receivedMessage.Payload), message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func mergeUserLists(channelUsers, chatUsers []uuid.UUID) []uuid.UUID {
	userSet := make(map[uuid.UUID]struct{})
	for _, user := range channelUsers {
		userSet[user] = struct{}{}
	}
	for _, user := range chatUsers {
		userSet[user] = struct{}{}
	}
	var mergedUsers []uuid.UUID
	for user := range userSet {
		mergedUsers = append(mergedUsers, user)
	}
	return mergedUsers
}
