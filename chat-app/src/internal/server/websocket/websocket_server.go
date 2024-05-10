package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/chat-app/src/internal/dto"
	"example.com/chat-app/src/internal/models"
	"example.com/chat-app/src/internal/repository"
	"example.com/chat-app/src/redis"
	"github.com/gorilla/websocket"
)

var ctx = context.Background()
var userIdXWsConnection = make(map[string]*websocket.Conn)
var broadcastChannel = make(chan *models.Message)
var filePathsChannel = make(chan *models.MessageIdXFileId)

type WebsocketService struct {
	Repository *repository.Repository
}

func New(r *repository.Repository) *WebsocketService {
	return &WebsocketService{
		Repository: r,
	}
}

func (ws *WebsocketService) serveWs(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occured while trying to connect to websocket server.")
	}

	authToken := r.Header.Get("Authorization")

	//Let's say that senderJWT is sent to AuthService which will return userId
	userId := authToken

	slog.Info("WebSocket Connection opened to user " + userId)

	userIdXWsConnection[userId] = wsConnection
	ws.readMessages(wsConnection)

	delete(userIdXWsConnection, userId)
	wsConnection.Close()
}

func (ws *WebsocketService) readMessages(wsConnection *websocket.Conn) {
	for {
		_, payload, err := wsConnection.ReadMessage()
		if err != nil {
			slog.Error("Error has occured while reading message", err)
			return
		}

		messageReq := dto.MessageRequest{}

		err = json.Unmarshal(payload, &messageReq)
		if err != nil {
			slog.Error("Error has occured while unmarshalling message", err)
			return
		}

		message, err := models.MapRequestToMessage(&messageReq)
		if err != nil {
			slog.Error("Error has occured while unmarshalling message", err)
			return
		}

		if messageReq.WithMedia {
			for {
				mf := <-filePathsChannel
				if mf.MessageId == message.Id {
					metadata := dto.Metadata{}
					metadata.FilePath = mf.FileId.String()
					message.Metadata = metadata
					break
				}
			}
		}

		//Let's say that senderJWT is sent to AuthService which will return userId
		slog.Info("SenderId:" + fmt.Sprintf("%d", message.SenderId) + "\tBody:" + message.Body + "\tChatRoomId:" + fmt.Sprintf("%d", message.ChatRoomId) + ":\tTime:" + fmt.Sprintf("%d", message.CreatedAt) + /*"\tMetadata:" + fmt.Sprintf("%v", message.Metadata) +*/ "\tWithMedia:" + fmt.Sprintf("%v", message.WithMedia))
		err = ws.Repository.SaveUserMessage(message)
		if err != nil {
			slog.Error("Error has occured while saving message", err)
			return
		}
		broadcastChannel <- message
	}
}

func (ws *WebsocketService) broadcast() {
	for {
		message := <-broadcastChannel
		slog.Info("New message from " + fmt.Sprintf("%d", message.SenderId) + ":\t" + message.Body)

		//Let's say that message.ChatRoomId is sent to ChatRoomMgmtService which will return List<UserId>
		userIds := []string{"1", "2"}
		for _, userId := range userIds {
			wsConnection, ok := userIdXWsConnection[userId]
			if ok {
				messageResp := models.MapMessageToResponse(message)
				err := wsConnection.WriteJSON(messageResp)
				if err != nil {
					slog.Error("Error has occured while sending message", err)
					wsConnection.Close()
					delete(userIdXWsConnection, userId)
				}
			} else {
				//NotificationService will listen 'notification' channel in Redis Pub/Sub
				redis.RedisClient.Publish(ctx, "notification", message)
			}
		}
	}
}

func (ws *WebsocketService) listenFileChannel() {
	var subscriber = redis.RedisClient.Subscribe(ctx, "file-loaded-channel")
	for {
		message, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			continue
		}
		mf := &models.MessageIdXFileId{}
		json.Unmarshal([]byte(message.String()), mf)
		filePathsChannel <- mf
	}
}

func (ws *WebsocketService) setupRoutes() {
	http.HandleFunc("/websocket", ws.serveWs)
}

func (ws *WebsocketService) SetupServer() {
	go ws.broadcast()
	go ws.listenFileChannel()
	ws.setupRoutes()
}
