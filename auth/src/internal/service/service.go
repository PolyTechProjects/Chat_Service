package service

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"example.com/main/src/internal/repository"
	"example.com/main/src/models"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepository *repository.AuthRepository
	jwtSecretKey   []byte
	keyFunc        func(token *jwt.Token) (interface{}, error)
}

func New(authRepository *repository.AuthRepository) *AuthService {
	jwtSecretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	}
	return &AuthService{AuthRepository: authRepository, jwtSecretKey: jwtSecretKey, keyFunc: keyFunc}
}

func (s *AuthService) Register(login string, username string, password string) (string, string, uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", uuid.Nil, err
	}
	user, err := models.New(login, username, string(hash))
	if err != nil {
		return "", "", uuid.Nil, err
	}
	err = s.AuthRepository.Save(user)
	if err != nil {
		return "", "", uuid.Nil, err
	}

	refreshTokenValueString := fmt.Sprintf("%v:%v:%v", user.Id, user.Login, time.Now().Unix())
	refreshTokenValue, err := bcrypt.GenerateFromPassword([]byte(refreshTokenValueString), bcrypt.DefaultCost)
	if err != nil {
		return "", "", uuid.Nil, err
	}
	refreshToken := models.NewRefreshToken(user.Id, string(refreshTokenValue))

	accessToken, err := s.generateAccessToken(user.Id, user.Name)
	if err != nil {
		return "", "", uuid.Nil, err
	}
	slog.Info(fmt.Sprintf("User %v registered", user.Id))
	return accessToken, refreshToken.Value, user.Id, nil
}

func (s *AuthService) Login(login string, password string) (string, string, error) {
	user, err := s.AuthRepository.FindByLogin(login)
	if err != nil {
		return "", "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(password))
	if err != nil {
		return "", "", err
	}

	refreshTokenValueString := fmt.Sprintf("%v:%v:%v", user.Id, user.Login, time.Now().Unix())
	slog.Debug(fmt.Sprintf("refreshTokenValueString: %v", refreshTokenValueString))
	refreshTokenValue, err := bcrypt.GenerateFromPassword([]byte(refreshTokenValueString), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	refreshToken := models.NewRefreshToken(user.Id, string(refreshTokenValue))
	slog.Debug(fmt.Sprintf("refreshToken: %v", refreshToken))

	accessToken, err := s.generateAccessToken(user.Id, user.Name)
	slog.Debug(fmt.Sprintf("accessToken: %v", accessToken))
	if err != nil {
		return "", "", err
	}
	slog.Info(fmt.Sprintf("User %v authenticated", user.Id))
	return accessToken, refreshToken.Value, nil
}

func (s *AuthService) Authorize(accessToken string, refreshToken string) (string, string, uuid.UUID, error) {
	var claims jwt.MapClaims
	_, err := jwt.ParseWithClaims(accessToken, &claims, s.keyFunc)
	if err != nil {
		return "", "", uuid.Nil, err
	}

	userId, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return "", "", uuid.Nil, err
	}

	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return "", "", uuid.Nil, err
	}

	if claims["exp"].(float64) < float64(time.Now().Unix()) {
		accessToken, err = s.refreshAccessToken(refreshToken, user.Id)
		if err != nil {
			return "", "", uuid.Nil, err
		}
	}

	return accessToken, refreshToken, user.Id, nil
}

func (s *AuthService) ExtractUserId(tokenString string) (string, error) {
	var claims jwt.MapClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, s.keyFunc)
	if err != nil {
		return "", err
	}
	return claims["sub"].(string), nil
}

func (s *AuthService) refreshAccessToken(refreshToken string, userId uuid.UUID) (string, error) {
	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return "", err
	}
	userRefreshToken, err := s.AuthRepository.FindTokenByUserId(user.Id)
	if err != nil {
		return "", err
	}
	if userRefreshToken.Value != refreshToken {
		return "", fmt.Errorf("invalid refresh token")
	}
	if userRefreshToken.ExpiredAt.Before(time.Now()) {
		return "", fmt.Errorf("refresh token expired")
	}
	token, err := s.generateAccessToken(user.Id, user.Name)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) generateAccessToken(userId uuid.UUID, userName string) (string, error) {
	payload := jwt.MapClaims{
		"sub":  userId,
		"name": userName,
		"exp":  time.Now().Add(time.Minute * 30).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString(s.jwtSecretKey)
}
