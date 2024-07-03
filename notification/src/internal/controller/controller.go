package controller

import (
	"encoding/json"
	"fmt"
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
	var bindDeviceToUserRequest dto.BindDeviceToUserRequest
	err := json.NewDecoder(r.Body).Decode(&bindDeviceToUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResp, err := nc.authClient.PerformAuthorize(r.Context(), r, bindDeviceToUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
}

func (nc *NotificationController) UnbindDeviceFromUserHandler(w http.ResponseWriter, r *http.Request) {
	var unbindDeviceFromUserRequest dto.UnbindDeviceFromUserRequest
	err := json.NewDecoder(r.Body).Decode(&unbindDeviceFromUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResp, err := nc.authClient.PerformAuthorize(r.Context(), r, unbindDeviceFromUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
}

func (nc *NotificationController) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var deleteUserRequest dto.DeleteUserRequest
	err := json.NewDecoder(r.Body).Decode(&deleteUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResp, err := nc.authClient.PerformAuthorize(r.Context(), r, deleteUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
}

func (nc *NotificationController) UpdateOldDeviceOnUserHandler(w http.ResponseWriter, r *http.Request) {
	var updateOldDeviceOnUserRequest dto.UpdateOldDeviceOnUserRequest
	err := json.NewDecoder(r.Body).Decode(&updateOldDeviceOnUserRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResp, err := nc.authClient.PerformAuthorize(r.Context(), r, updateOldDeviceOnUserRequest.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
}
