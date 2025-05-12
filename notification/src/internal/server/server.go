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

func (n *NotificationHttpServer) StartServer() {
	http.HandleFunc("DELETE /api/v1/notifications/subscription/{chatId}", n.notificationController.UnsubscribeHandler)
}
