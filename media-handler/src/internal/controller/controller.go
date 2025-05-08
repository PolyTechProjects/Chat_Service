package controller

import (
	"net/http"
	"net/url"

	"example.com/media/src/internal/client"
	"example.com/media/src/internal/service"
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
	_, err := m.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = m.mediaHandlerService.UploadMedia(file, fileHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func (m *MediaHandlerController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
	_, err := m.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("mediaId") {
		http.Error(w, "URL query params are invalid", http.StatusBadRequest)
	}
	mediaId := params.Get("mediaId")
	id, err := uuid.Parse(mediaId)
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
	w.Write(res)
}

func (m *MediaHandlerController) DeleteMediaHandler(w http.ResponseWriter, r *http.Request) {
	_, err := m.authClient.PerformAuthorize(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !params.Has("mediaId") {
		http.Error(w, "URL query params are invalid", http.StatusBadRequest)
	}
	mediaId := params.Get("mediaId")
	id, err := uuid.Parse(mediaId)
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
}
