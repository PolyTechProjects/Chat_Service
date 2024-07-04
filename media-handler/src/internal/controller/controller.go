package controller

import (
	"fmt"
	"net/http"
	"strings"

	"example.com/media-handler/src/internal/client"
	"example.com/media-handler/src/internal/service"
	"github.com/google/uuid"
)

type MediaHandlerController struct {
	mediaHandlerService *service.MediaHandlerService
	authClient          *client.AuthGRPCClient
}

func New(mediaHandlerService *service.MediaHandlerService, authClient *client.AuthGRPCClient) *MediaHandlerController {
	return &MediaHandlerController{mediaHandlerService: mediaHandlerService, authClient: authClient}
}

func (m *MediaHandlerController) UploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := m.authClient.PerformAuthorize(r.Context(), r, r.Header.Get("UserId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	header := r.Header.Get("MessageId")
	messageId, err := uuid.Parse(header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = m.mediaHandlerService.UploadMedia(messageId, file, fileHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}

func (m *MediaHandlerController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := m.authClient.PerformAuthorize(r.Context(), r, r.Header.Get("UserId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(strings.Split(r.URL.Path, "/")[2])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := m.mediaHandlerService.GetMedia(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
	w.Write(res)
}

func (m *MediaHandlerController) DeleteMediaHandler(w http.ResponseWriter, r *http.Request) {
	authResp, err := m.authClient.PerformAuthorize(r.Context(), r, r.Header.Get("UserId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(strings.Split(r.URL.Path, "/")[2])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = m.mediaHandlerService.DeleteMedia(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", authResp.AccessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", authResp.RefreshToken))
}
