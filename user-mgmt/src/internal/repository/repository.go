package repository

import (
	"example.com/user-mgmt/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type UserMgmtRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *UserMgmtRepository {
	return &UserMgmtRepository{db: db}
}

func (r *UserMgmtRepository) GetUser(userId uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserMgmtRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserMgmtRepository) DeleteUser(userId uuid.UUID) error {
	var user models.User
	return r.db.Where("id = ?", userId).Delete(user).Error
}
