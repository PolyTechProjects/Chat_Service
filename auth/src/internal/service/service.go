package service

import (
	"os"

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

func (s *AuthService) Register(login string, username string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user, err := models.New(login, username, string(hash))
	if err != nil {
		return "", err
	}
	err = s.AuthRepository.Save(user)
	if err != nil {
		return "", err
	}

	payload := jwt.MapClaims{
		"sub":  user.Id,
		"name": user.Name,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString(s.jwtSecretKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) Login(login string, password string) (string, error) {
	user, err := s.AuthRepository.FindByLogin(login)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(password))
	if err != nil {
		return "", err
	}

	payload := jwt.MapClaims{
		"sub":  user.Id,
		"name": user.Name,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString(s.jwtSecretKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) Authorize(tokenString string) error {
	var claims jwt.MapClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, s.keyFunc)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return err
	}

	_, err = s.AuthRepository.FindById(userId)
	if err != nil {
		return err
	}

	return nil
}
