package repository

import (
	"log/slog"

	"example.com/users/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type UsersRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Save(user *models.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		slog.Error("RepositorySave failed: " + err.Error())
		return err
	}
	return nil
}

func (r *UsersRepository) DeleteById(userId uuid.UUID) error {
	var user models.User
	err := r.db.Where("id = ?", userId).Delete(user).Error
	if err != nil {
		slog.Error("RepositoryDeleteById failed: " + err.Error())
		return err
	}
	return nil
}

func (r *UsersRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	if err != nil {
		slog.Error("RepositoryGetAll failed: " + err.Error())
		return nil, err
	}
	return users, nil
}

func (r *UsersRepository) GetByIds(userIds []uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN ?", userIds).Find(&users).Error
	if err != nil {
		slog.Error("RepositoryGetByIds failed: " + err.Error())
		return nil, err
	}
	return users, nil
}

func (r *UsersRepository) GetById(userId uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.Where("id = ?", userId).First(user).Error
	if err != nil {
		slog.Error("RepositoryGetById failed: " + err.Error())
		return nil, err
	}
	return user, nil
}

func (r *UsersRepository) GetByProfileLink(profileLink string) (*models.User, error) {
	user := &models.User{}
	err := r.db.Where("profile_link = ?", profileLink).First(user).Error
	if err != nil {
		slog.Error("RepositoryGetById failed: " + err.Error())
		return nil, err
	}
	return user, nil
}

func (r *UsersRepository) Update(user *models.User) (*models.User, error) {
	r.db.Begin()
	err := r.db.Save(user).Error
	if err != nil {
		slog.Error("RepositoryUpdate failed: " + err.Error())
		r.db.Rollback()
		return nil, err
	}
	r.db.Commit()
	return user, nil
}
