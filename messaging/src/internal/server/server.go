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

func (h *HttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/messaging/history/direct/{chatId}", h.messageHistoryController.GetDirectHistoryHandler)
	mux.HandleFunc("GET /api/v1/messaging/history/{chatId}", h.messageHistoryController.GetHistoryHandler)
	//http.HandleFunc("DELETE /api/v1/messaging/history", h.messageHistoryController.DeleteHistoryHandler)
	mux.HandleFunc("DELETE /api/v1/messaging/history/{messageId}", h.messageHistoryController.DeleteMessageHandler)
	mux.HandleFunc("PUT /api/v1/messaging/history", h.messageHistoryController.EditMessageHandler)
	mux.HandleFunc("/api/v1/messaging/ws/{chatId}", h.websocketController.SendMessageHandler)
	mux.HandleFunc("/api/v1/messaging/ws/direct/{chatId}", h.websocketController.SendDirectMessageHandler)
}

func (h *HttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
