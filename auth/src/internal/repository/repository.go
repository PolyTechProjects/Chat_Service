package repository

import (
	"example.com/main/src/models"
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

func (r *AuthRepository) FindByLogin(login string) (*models.User, error) {
	var user models.User
	err := r.db.Where("login = ?", login).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
