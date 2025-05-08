package service

import (
	"log/slog"

	"example.com/user_mgmt/src/internal/dto"
	"example.com/user_mgmt/src/internal/repository"
	"example.com/user_mgmt/src/models"
	"github.com/google/uuid"
)

type UserMgmtService struct {
	UserMgmtRepository *repository.UserMgmtRepository
}

func New(repository *repository.UserMgmtRepository) *UserMgmtService {
	return &UserMgmtService{
		UserMgmtRepository: repository,
	}
}

func (s *UserMgmtService) CreateUser(accountCreatedEvent *dto.AccountCreatedEvent) error {
	user := models.New(
		uuid.MustParse(accountCreatedEvent.UserId),
		accountCreatedEvent.Username,
		accountCreatedEvent.Firstname,
		accountCreatedEvent.Lastname,
	)
	err := s.UserMgmtRepository.Save(user)
	if err != nil {
		return err
	}
	slog.Debug("Created user: " + accountCreatedEvent.UserId)
	return err
}

func (s *UserMgmtService) DeleteUser(accountDeletedEvent *dto.AccountDeletedEvent) error {
	err := s.UserMgmtRepository.DeleteById(uuid.MustParse(accountDeletedEvent.UserId))
	if err != nil {
		return err
	}
	slog.Debug("Deleted user: " + accountDeletedEvent.UserId)
	return nil
}

func (s *UserMgmtService) GetUsers() (*dto.UsersResponse, error) {
	userEntities, err := s.UserMgmtRepository.GetAll()
	if err != nil {
		return nil, err
	}
	users := &dto.UsersResponse{
		Users: make([]dto.UserResponse, len(userEntities)),
	}
	for i, userEntity := range userEntities {
		users.Users[i] = dto.UserResponse{
			UserId:      userEntity.Id.String(),
			Name:        userEntity.Name,
			Firstname:   userEntity.Firstname,
			Lastname:    userEntity.Lastname,
			ProfilePic:  userEntity.ProfilePic,
			ProfileLink: userEntity.ProfileLink,
			Description: userEntity.Description,
		}
	}
	return users, nil
}

func (s *UserMgmtService) GetUser(profileLink string) (*dto.UserResponse, error) {
	user, err := s.UserMgmtRepository.GetByProfileLink(profileLink)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Firstname:   user.Firstname,
		Lastname:    user.Lastname,
		ProfilePic:  user.ProfilePic,
		ProfileLink: user.ProfileLink,
		Description: user.Description,
	}, nil
}

func (s *UserMgmtService) UpdateUser(profileLink string, updateUserRequest *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.UserMgmtRepository.GetByProfileLink(profileLink)
	if err != nil {
		return nil, err
	}
	user.Name = updateUserRequest.Name
	user.Firstname = updateUserRequest.Firstname
	user.Lastname = updateUserRequest.Lastname
	user.ProfileLink = updateUserRequest.ProfileLink
	user.Description = updateUserRequest.Description
	updatedUser, err := s.UserMgmtRepository.Update(user)
	return &dto.UserResponse{
		UserId:      updatedUser.Id.String(),
		Name:        updatedUser.Name,
		Firstname:   updatedUser.Firstname,
		Lastname:    updatedUser.Lastname,
		ProfilePic:  updatedUser.ProfilePic,
		ProfileLink: updatedUser.ProfileLink,
		Description: updatedUser.Description,
	}, err
}
