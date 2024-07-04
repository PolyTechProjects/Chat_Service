package server

import (
	"net/http"

	"example.com/chat-app/src/internal/controller"
)

type HttpServer struct {
	messageHistoryController *controller.MessageHistoryController
	websocketController      *controller.WebsocketController
}

func NewHttpServer(messageHistoryController *controller.MessageHistoryController, websocketController *controller.WebsocketController) *HttpServer {
	return &HttpServer{
		messageHistoryController: messageHistoryController,
		websocketController:      websocketController,
	}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("GET /{chatRoomId}/history", h.messageHistoryController.GetHistoryHandler)
	http.HandleFunc("/websocket/channel", h.websocketController.SendMessageInChannelHandler)
	http.HandleFunc("/websocket/chat", h.websocketController.SendMessageInChatRoomHandler)
	go h.websocketController.StartBroadcastingToChatRooms()
	go h.websocketController.StartBroadcastingToChannels()
	go h.websocketController.StartListeningFileChannel()
}
