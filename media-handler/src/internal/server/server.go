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

func (h *HttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/media/uploads", h.mediaHandlerController.UploadMediaHandler)
	mux.HandleFunc("GET /api/v1/media/uploads", h.mediaHandlerController.GetMediaHandler)
	mux.HandleFunc("DELETE /api/v1/media/uploads", h.mediaHandlerController.DeleteMediaHandler)
}

func (h *HttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
