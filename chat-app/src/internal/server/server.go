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
	http.HandleFunc("/websocket", h.websocketController.SendMessageHandler)
	go h.websocketController.StartBroadcasting()
	go h.websocketController.StartListeningFileChannel()
}
