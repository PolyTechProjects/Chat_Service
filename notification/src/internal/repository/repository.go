package repository

import (
	"example.com/notification/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type UserIdXDeviceTokenRepository struct {
	db *gorm.DB
}

func NewUserIdXDeviceTokenRepository(db *gorm.DB) *UserIdXDeviceTokenRepository {
	return &UserIdXDeviceTokenRepository{db: db}
}

func (udtr *UserIdXDeviceTokenRepository) GetByUserId(userId uuid.UUID) ([]*models.UserIdXDeviceToken, error) {
	var userIdXDeviceTokens []*models.UserIdXDeviceToken
	err := udtr.db.Where("user_id = ?", userId).Find(&userIdXDeviceTokens).Error
	if err != nil {
		return nil, err
	}
	return userIdXDeviceTokens, nil
}

func (udtr *UserIdXDeviceTokenRepository) BindDeviceTokenToUser(userIdXDeviceToken *models.UserIdXDeviceToken) error {
	return udtr.db.Create(userIdXDeviceToken).Error
}

func (udtr *UserIdXDeviceTokenRepository) UnbindDeviceTokenFromUser(userId uuid.UUID, deviceToken string) error {
	return udtr.db.Where("user_id = ? AND device_token = ?", userId, deviceToken).Delete(&models.UserIdXDeviceToken{}).Error
}

func (udtr *UserIdXDeviceTokenRepository) DeleteUser(userId uuid.UUID) error {
	return udtr.db.Where("user_id = ?", userId).Delete(&models.UserIdXDeviceToken{}).Error
}

func (udtr *UserIdXDeviceTokenRepository) UpdateDeviceTokensByUserId(userId uuid.UUID, oldDeviceToken string, newDeviceToken string) error {
	return udtr.db.Where("user_id = ? AND device_token = ?", userId, oldDeviceToken).Update("device_token", newDeviceToken).Error
}
