package server

import (
	"net/http"

	"example.com/messaging/src/internal/controller"
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
	http.HandleFunc("GET /api/v1/messaging/history/direct/{userId}", h.messageHistoryController.GetDirectHistoryHandler)
	http.HandleFunc("GET /api/v1/messaging/history/{chatId}", h.messageHistoryController.GetHistoryHandler)
	http.HandleFunc("/api/v1/messaging/ws/{chatId}", h.websocketController.SendMessageHandler)
	http.HandleFunc("/api/v1/messaging/ws/direct/{userId}", h.websocketController.SendDirectMessageHandler)
}
