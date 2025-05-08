package server

import (
	"net/http"

	"example.com/user_mgmt/src/internal/controller"
)

type HttpServer struct {
	userMgmtController *controller.UserMgmtController
}

func NewHttpServer(userMgmtController *controller.UserMgmtController) *HttpServer {
	return &HttpServer{userMgmtController: userMgmtController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("PUT /api/v1/users/profiles/{profileLink}", h.userMgmtController.UpdateProfileHandler)
	http.HandleFunc("GET /api/v1/users/profiles", h.userMgmtController.GetProfilesHandler)
}
