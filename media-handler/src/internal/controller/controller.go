package controller

import (
	"net/http"
	"strings"

	"example.com/media-handler/src/internal/service"
	"github.com/google/uuid"
)

type MediaHandlerController struct {
	mediaHandlerService *service.MediaHandlerService
}

func New(mediaHandlerService *service.MediaHandlerService) *MediaHandlerController {
	return &MediaHandlerController{mediaHandlerService: mediaHandlerService}
}

func (m *MediaHandlerController) UploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("MessageId")
	messageId, err := uuid.Parse(header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = m.mediaHandlerService.UploadMedia(messageId, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (m *MediaHandlerController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write(res)
}

func (m *MediaHandlerController) DeleteMediaHandler(w http.ResponseWriter, r *http.Request) {
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
}
