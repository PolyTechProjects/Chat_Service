package server

import (
	"net/http"

	"example.com/notification/src/internal/controller"
)

type NotificationHttpServer struct {
	notificationController *controller.NotificationController
}

func NewNotificationHttpServer(notificationController *controller.NotificationController) *NotificationHttpServer {
	return &NotificationHttpServer{
		notificationController: notificationController,
	}
}

func (n *NotificationHttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("DELETE /api/v1/notifications/subscription/{chatId}", n.notificationController.UnsubscribeHandler)
}

func (h *NotificationHttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
