package server

import (
	"net/http"

	"example.com/users/src/internal/controller"
)

type HttpServer struct {
	UsersController *controller.UsersController
}

func NewHttpServer(UsersController *controller.UsersController) *HttpServer {
	return &HttpServer{UsersController: UsersController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("PUT /api/v1/users/profiles/{profileLink}", h.UsersController.UpdateProfileHandler)
	http.HandleFunc("GET /api/v1/users/profiles", h.UsersController.GetProfilesHandler)
}
