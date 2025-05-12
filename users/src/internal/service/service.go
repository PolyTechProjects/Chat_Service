package service

import (
	"log/slog"

	"example.com/users/src/internal/dto"
	"example.com/users/src/internal/repository"
	"example.com/users/src/models"
	"github.com/google/uuid"
)

type UsersService struct {
	UsersRepository *repository.UsersRepository
}

func New(repository *repository.UsersRepository) *UsersService {
	return &UsersService{
		UsersRepository: repository,
	}
}

func (s *UsersService) CreateUser(accountCreatedEvent *dto.AccountCreatedEvent) error {
	user := models.New(
		uuid.MustParse(accountCreatedEvent.UserId),
		accountCreatedEvent.Username,
		accountCreatedEvent.Firstname,
		accountCreatedEvent.Lastname,
	)
	err := s.UsersRepository.Save(user)
	if err != nil {
		return err
	}
	slog.Debug("Created user: " + accountCreatedEvent.UserId)
	return err
}

func (s *UsersService) DeleteUser(accountDeletedEvent *dto.AccountDeletedEvent) error {
	err := s.UsersRepository.DeleteById(uuid.MustParse(accountDeletedEvent.UserId))
	if err != nil {
		return err
	}
	slog.Debug("Deleted user: " + accountDeletedEvent.UserId)
	return nil
}

func (s *UsersService) GetUsers() (*dto.UsersResponse, error) {
	userEntities, err := s.UsersRepository.GetAll()
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

func (s *UsersService) GetUsersByIds(userIds []uuid.UUID) (*dto.UsersResponse, error) {
	users := &dto.UsersResponse{
		Users: make([]dto.UserResponse, len(userIds)),
	}
	for i, userId := range userIds {
		user, err := s.UsersRepository.GetById(userId)
		if err != nil {
			return nil, err
		}
		users.Users[i] = dto.UserResponse{
			UserId:      user.Id.String(),
			Name:        user.Name,
			Firstname:   user.Firstname,
			Lastname:    user.Lastname,
			ProfilePic:  user.ProfilePic,
			ProfileLink: user.ProfileLink,
			Description: user.Description,
		}
	}
	return users, nil
}

func (s *UsersService) GetUser(profileLink string) (*dto.UserResponse, error) {
	user, err := s.UsersRepository.GetByProfileLink(profileLink)
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

func (s *UsersService) UpdateUser(profileLink string, updateUserRequest *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.UsersRepository.GetByProfileLink(profileLink)
	if err != nil {
		return nil, err
	}
	user.Name = updateUserRequest.Name
	user.Firstname = updateUserRequest.Firstname
	user.Lastname = updateUserRequest.Lastname
	user.ProfileLink = updateUserRequest.ProfileLink
	user.Description = updateUserRequest.Description
	updatedUser, err := s.UsersRepository.Update(user)
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
