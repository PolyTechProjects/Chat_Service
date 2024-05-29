package repository

import (
	"example.com/chat-management/src/internal/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) SaveChat(chat models.Chat) error {
	return r.db.Create(&chat).Error
}

func (r *ChatRepository) DeleteChat(chatID uuid.UUID) error {
	return r.db.Where("id = ?", chatID).Delete(&models.Chat{}).Error
}

func (r *ChatRepository) UpdateChat(chatID uuid.UUID, name, description string) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error
}

func (r *ChatRepository) AddUserToChat(chatID, userID uuid.UUID) error {
	userChat := models.UserChat{
		ChatID: chatID,
		UserID: userID,
	}
	return r.db.Create(&userChat).Error
}

func (r *ChatRepository) RemoveUserFromChat(chatID, userID uuid.UUID) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&models.UserChat{}).Error
}

func (r *ChatRepository) AddAdmin(chatID, userID uuid.UUID) error {
	admin := models.Admin{
		ChatID: chatID,
		UserID: userID,
	}
	return r.db.Create(&admin).Error
}

func (r *ChatRepository) RemoveAdmin(chatID, userID uuid.UUID) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&models.Admin{}).Error
}

func (r *ChatRepository) IsAdmin(chatID, userID uuid.UUID) (bool, error) {
	var admin models.Admin
	err := r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ChatRepository) IsMember(chatID, userID uuid.UUID) (bool, error) {
	var userChat models.UserChat
	err := r.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&userChat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
