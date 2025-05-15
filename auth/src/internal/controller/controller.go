package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"example.com/main/src/internal/dto"
	"example.com/main/src/internal/service"
	"github.com/google/uuid"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
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
	response, err := json.Marshal(&dto.LoginResponse{
		AccessToken: accessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly; Path=/api/v1/auth/refresh", refreshToken))
	w.Write(response)
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

func (a *AuthController) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	segments := strings.Split(path, "/")
	userId, err := uuid.Parse(segments[len(segments)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		http.Error(w, "no access token", http.StatusUnauthorized)
		return
	}
	refreshToken, err := r.Cookie("X-Refresh-Token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	err = a.authService.DeleteAccount(userId, accessToken, refreshToken.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
}

func (a *AuthController) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("RefreshHandler")
	refreshTokenCookie, err := r.Cookie("X-Refresh-Token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	refreshToken := refreshTokenCookie.Value
	if refreshToken == "" {
		http.Error(w, "no refresh token", http.StatusUnauthorized)
		return
	}
	accessToken, refreshToken, err := a.authService.RefreshTokens(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	response, err := json.Marshal(&dto.LoginResponse{
		AccessToken: accessToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Set-Cookie", fmt.Sprintf("X-Refresh-Token=%s; HttpOnly; SameSite=None; Path=/api/v1/auth/refresh", refreshToken))
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *AuthController) MeHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if accessToken == "" {
		http.Error(w, "no access token", http.StatusUnauthorized)
		return
	}
	userId, err := a.authService.ExtractUserId(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	response, err := json.Marshal(&dto.MeResponse{UserId: uuid.MustParse(userId)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
