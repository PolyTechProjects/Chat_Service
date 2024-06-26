package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"example.com/user-mgmt/src/internal/client"
	"example.com/user-mgmt/src/internal/dto"
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
	mediaResp, err := c.mediaHandlerClient.PerformStoreImage(r.Context(), authResp.AccessToken, file, fileHeader.Filename)
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

func (c *UserMgmtController) InfoUpdateHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := c.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	dto := dto.UpdateInfoRequest{}
	err = json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(dto.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := c.userMgmtService.UpdateUser(userId, dto.Name, dto.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
	w.Write(resp)
}

func (c *UserMgmtController) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := c.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("userId") {
		http.Error(w, "URL query params are invalid", http.StatusBadRequest)
	}
	userId, err := uuid.Parse(params.Get("userId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := c.userMgmtService.GetUser(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
	w.Write(resp)
}

func (c *UserMgmtController) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := c.authClient.PerformAuthorize(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	dto := dto.UpdateInfoRequest{}
	err = json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(dto.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = c.userMgmtService.DeleteUser(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=Bearer %s; X-Refresh-Token=%s", authResp.AccessToken, authResp.RefreshToken))
}
