package controller

import (
	"net/http"
	"strings"

	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/service"
	"github.com/google/uuid"
)

type NotificationController struct {
	notificationService *service.NotificationService
	authClient          *client.AuthGRPCClient
}

func NewNotificationController(notificationService *service.NotificationService, authClient *client.AuthGRPCClient) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		authClient:          authClient,
	}
}

func (c *NotificationController) UnsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	chatId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = c.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	extractResp, err := c.authClient.PerformExtractUserId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := uuid.Parse(extractResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.notificationService.Unsubscribe(chatId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
