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

func (r *AuthRepository) DeleteById(id uuid.UUID) error {
	err := r.db.Debug().Where("id = ?", id).Delete(&models.User{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *AuthRepository) FindByKeycloakId(keycloakId uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.Where("keycloak_id = ?", keycloakId).Find(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *AuthRepository) FindUserVerificationById(userId uuid.UUID) (*models.UserVerification, error) {
	userVerification := &models.UserVerification{}
	err := r.db.Where("user_id = ?", userId).Last(userVerification).Error
	if err != nil {
		return nil, err
	}
	return userVerification, nil
}

func (r *AuthRepository) StoreUserVerification(userVerification *models.UserVerification) (*models.UserVerification, error) {
	res := r.db.Save(userVerification)
	if res.Error != nil {
		return nil, res.Error
	}
	return res.Value.(*models.UserVerification), nil
}

func (r *AuthRepository) FindUserVerificationByToken(verificationToken string) (*models.UserVerification, error) {
	userVerification := &models.UserVerification{}
	err := r.db.Where("verification_token = ?", verificationToken).Find(userVerification).Error
	if err != nil {
		return nil, err
	}
	return userVerification, nil
}
