package server

import (
	"net/http"

	"example.com/media-handler/src/internal/controller"
)

type HttpServer struct {
	mediaHandlerController *controller.MediaHandlerController
}

func New(mediaHandlerController *controller.MediaHandlerController) *HttpServer {
	return &HttpServer{mediaHandlerController: mediaHandlerController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /upload", h.mediaHandlerController.UploadMediaHandler)
	http.HandleFunc("GET /uploads/{id}", h.mediaHandlerController.GetMediaHandler)
	http.HandleFunc("DELETE /delete/{id}", h.mediaHandlerController.DeleteMediaHandler)
}
