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
	http.HandleFunc("POST /api/v1/media/uploads", h.mediaHandlerController.UploadMediaHandler)
	http.HandleFunc("GET /api/v1/media/uploads", h.mediaHandlerController.GetMediaHandler)
	http.HandleFunc("DELETE /api/v1/media/uploads", h.mediaHandlerController.DeleteMediaHandler)
}
