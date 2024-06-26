package controller

import (
	"encoding/json"
	"net/http"

	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/dto"
	"example.com/notification/src/internal/service"
	"github.com/google/uuid"
)

type NotificationController struct {
	notificationService *service.NotificationService
	authClient          *client.AuthClient
	userMgmtClient      *client.UserMgmtClient
}

func NewNotificationController(notificationService *service.NotificationService, authClient *client.AuthClient, userMgmtClient *client.UserMgmtClient) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		authClient:          authClient,
		userMgmtClient:      userMgmtClient,
	}
}

func (nc *NotificationController) BindDeviceToUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := nc.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var bindDeviceToUserRequest dto.BindDeviceToUserRequest
	err = json.NewDecoder(r.Body).Decode(&bindDeviceToUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(bindDeviceToUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = nc.notificationService.BindDeviceToUser(userId, bindDeviceToUserRequest.DeviceToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (nc *NotificationController) UnbindDeviceFromUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := nc.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var unbindDeviceFromUserRequest dto.UnbindDeviceFromUserRequest
	err = json.NewDecoder(r.Body).Decode(&unbindDeviceFromUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(unbindDeviceFromUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = nc.notificationService.UnbindDeviceFromUser(userId, unbindDeviceFromUserRequest.DeviceToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (nc *NotificationController) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := nc.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var deleteUserRequest dto.DeleteUserRequest
	err = json.NewDecoder(r.Body).Decode(&deleteUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(deleteUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = nc.notificationService.DeleteUser(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (nc *NotificationController) UpdateOldDeviceOnUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := nc.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var updateOldDeviceOnUserRequest dto.UpdateOldDeviceOnUserRequest
	err = json.NewDecoder(r.Body).Decode(&updateOldDeviceOnUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(updateOldDeviceOnUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = nc.notificationService.UpdateOldDeviceOnUser(userId, updateOldDeviceOnUserRequest.OldDeviceToken, updateOldDeviceOnUserRequest.NewDeviceToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
