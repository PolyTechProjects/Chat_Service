package server

import (
	"net/http"

	"example.com/media/src/internal/controller"
)

type HttpServer struct {
	mediaHandlerController *controller.MediaHandlerController
}

func NewHttpServer(mediaHandlerController *controller.MediaHandlerController) *HttpServer {
	return &HttpServer{mediaHandlerController: mediaHandlerController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /uploads", h.mediaHandlerController.UploadMediaHandler)
	http.HandleFunc("GET /uploads", h.mediaHandlerController.GetMediaHandler)
	http.HandleFunc("DELETE /uploads", h.mediaHandlerController.DeleteMediaHandler)
}
