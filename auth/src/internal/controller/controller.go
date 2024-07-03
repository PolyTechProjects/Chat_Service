package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"example.com/main/src/internal/client"
	"example.com/main/src/internal/dto"
	"example.com/main/src/internal/service"
)

type AuthController struct {
	authService    *service.AuthService
	userMgmtClient *client.UserMgmtGRPCClient
}

func NewAuthController(authService *service.AuthService, userMgmtClient *client.UserMgmtGRPCClient) *AuthController {
	return &AuthController{
		authService:    authService,
		userMgmtClient: userMgmtClient,
	}
}

func (a *AuthController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	accessToken, refreshToken, userId, err := a.authService.Register(req.Login, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = a.userMgmtClient.PerformAddUser(r.Context(), userId.String(), req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	register := dto.RegisterResponse{
		UserId: userId.String(),
	}
	registerResp, err := json.Marshal(register)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", accessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", refreshToken))
	w.Write(registerResp)
}

func (a *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	accessToken, refreshToken, err := a.authService.Login(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly", accessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly", refreshToken))
}
