package service

import (
	"example.com/user-mgmt/src/internal/repository"
	"example.com/user-mgmt/src/models"
	"github.com/google/uuid"
)

type UserMgmtService struct {
	Repository *repository.UserMgmtRepository
}

func New(repository *repository.UserMgmtRepository) *UserMgmtService {
	return &UserMgmtService{Repository: repository}
}

func (s *UserMgmtService) CreateUser(userId uuid.UUID, name string) (*models.User, error) {
	user := models.New(userId, name)
	err := s.Repository.InsertUser(user)
	return user, err
}

func (s *UserMgmtService) UpdateUser(userId uuid.UUID, name string, description string) (*models.User, error) {
	user, err := s.GetUser(userId)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Description = description
	err = s.Repository.UpdateUser(user)
	return user, err
}

func (s *UserMgmtService) DeleteUser(userId uuid.UUID) error {
	return s.Repository.DeleteUser(userId)
}

func (s *UserMgmtService) GetUser(userId uuid.UUID) (*models.User, error) {
	user, err := s.Repository.GetUser(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}
