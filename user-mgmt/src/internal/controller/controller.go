package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"example.com/user-mgmt/src/internal/client"
	"example.com/user-mgmt/src/internal/service"
	"github.com/google/uuid"
)

type UserMgmtController struct {
	userMgmtService    *service.UserMgmtService
	authClient         *client.AuthGRPCClient
	mediaHandlerClient *client.MediaHandlerGRPCClient
}

func New(userMgmtService *service.UserMgmtService, authClient *client.AuthGRPCClient, mediaHandlerClient *client.MediaHandlerGRPCClient) *UserMgmtController {
	return &UserMgmtController{userMgmtService: userMgmtService, authClient: authClient, mediaHandlerClient: mediaHandlerClient}
}

func (c *UserMgmtController) UpdateAvatarHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	authResp, err := c.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.Info("Storing Image")
	mediaResp, err := c.mediaHandlerClient.PerformStoreImage(r.Context(), token, file, fileHeader.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userId, err := uuid.Parse(authResp.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := c.userMgmtService.UpdateAvatar(userId, mediaResp.FileId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
