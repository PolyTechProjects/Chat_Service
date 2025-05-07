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
	user, err := a.authService.Register(req.Login, req.Username, req.Password, req.Firstname, req.Lastname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	register := dto.RegisterResponse{
		UserId:    user.Id.String(),
		Login:     user.Login,
		Username:  user.Name,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
	}
	registerResp, err := json.Marshal(register)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Add("Set-Cookie", fmt.Sprintf("Authorization=%s; HttpOnly; Secure", accessToken))
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly; Secure", refreshToken))
}

func (a *AuthController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("X-Refresh-Token")
	if refreshToken == "" {
		http.Error(w, "no refresh token", http.StatusUnauthorized)
		return
	}
	err := a.authService.Logout(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
}
