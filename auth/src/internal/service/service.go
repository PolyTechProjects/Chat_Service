package service

import (
	"example.com/main/src/internal/client"
	"example.com/main/src/internal/dto"
	"example.com/main/src/internal/repository"
	"example.com/main/src/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	AuthRepository *repository.AuthRepository
	KeycloakClient *client.KeycloakClient
	RedisClient    *client.RedisClient
}

func New(authRepository *repository.AuthRepository, keycloakClient *client.KeycloakClient, redisClient *client.RedisClient) *AuthService {
	return &AuthService{
		AuthRepository: authRepository,
		KeycloakClient: keycloakClient,
		RedisClient:    redisClient,
	}
}

func (s *AuthService) Register(login string, username string, password string, firstname string, lastname string) (*models.User, error) {
	user, err := models.New(login, username, password, firstname, lastname)
	if err != nil {
		return nil, err
	}
	err = s.KeycloakClient.RegisterUser(user)
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Pass = string(hash)
	err = s.AuthRepository.Save(user)
	if err != nil {
		return nil, err
	}
	accountCreatedEvent := &dto.AccountCreatedEvent{
		UserId:    user.Id.String(),
		Login:     user.Login,
		Username:  user.Name,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	}
	s.RedisClient.SendToCreateAccountChannel(accountCreatedEvent)

	return user, nil
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

	return s.KeycloakClient.LoginUser(login, password)
}

func (s *AuthService) Authorize(accessToken string) error {
	return s.KeycloakClient.AuthroizeUser(accessToken)
}

func (s *AuthService) RefreshTokens(refreshToken string) (string, string, error) {
	return s.KeycloakClient.RefreshTokens(refreshToken)
}

func (s *AuthService) Logout(refreshToken string) error {
	return s.KeycloakClient.RevokeTokens(refreshToken)
}

func (s *AuthService) DeleteAccount(userId uuid.UUID, accessToken string, refreshToken string) error {
	user, err := s.AuthRepository.FindById(userId)
	if err != nil {
		return err
	}
	err = s.KeycloakClient.DeleteAccount(user.KeycloakId, accessToken, refreshToken)
	if err != nil {
		return err
	}
	err = s.AuthRepository.DeleteById(userId)
	if err != nil {
		return err
	}
	accountDeletedEvent := &dto.AccountDeletedEvent{
		UserId: user.Id.String(),
	}
	s.RedisClient.SendToDeleteAccountChannel(accountDeletedEvent)

	return nil
}
