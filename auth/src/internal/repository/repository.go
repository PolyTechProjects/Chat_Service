package repository

import (
	"example.com/main/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Save(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) FindById(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindByLogin(login string) (*models.User, error) {
	var user models.User
	err := r.db.Where("login = ?", login).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindTokenByUserId(userId uuid.UUID) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Where("user_id = ?", userId).Find(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}
