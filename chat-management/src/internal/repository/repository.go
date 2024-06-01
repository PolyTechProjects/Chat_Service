package repository

import (
	"example.com/chat-management/src/internal/dto"
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

func (r *ChatRepository) UpdateChat(request dto.UpdateChatRequest) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", request.ChatID).Updates(map[string]interface{}{
		"name":        request.Name,
		"description": request.Description,
	}).Error
}

func (r *ChatRepository) AddUserToChat(request dto.UserChatRequest) error {
	userChat := models.UserChat{
		ChatID: request.ChatID,
		UserID: request.UserID,
	}
	return r.db.Create(&userChat).Error
}

func (r *ChatRepository) RemoveUserFromChat(request dto.UserChatRequest) error {
	return r.db.Where("chat_id = ? AND user_id = ?", request.ChatID, request.UserID).Delete(&models.UserChat{}).Error
}

func (r *ChatRepository) AddAdmin(request dto.AdminRequest) error {
	admin := models.Admin{
		ChatID: request.ChatID,
		UserID: request.UserID,
	}
	return r.db.Create(&admin).Error
}

func (r *ChatRepository) RemoveAdmin(request dto.AdminRequest) error {
	return r.db.Where("chat_id = ? AND user_id = ?", request.ChatID, request.UserID).Delete(&models.Admin{}).Error
}

func (r *ChatRepository) IsAdmin(request dto.AdminRequest) (bool, error) {
	var admin models.Admin
	err := r.db.Where("chat_id = ? AND user_id = ?", request.ChatID, request.UserID).First(&admin).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ChatRepository) IsMember(request dto.UserChatRequest) (bool, error) {
	var userChat models.UserChat
	err := r.db.Where("chat_id = ? AND user_id = ?", request.ChatID, request.UserID).First(&userChat).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ChatRepository) GetChatUsers(chatID uuid.UUID) ([]string, error) {
	var userChats []models.UserChat
	err := r.db.Where("chat_id = ?", chatID).Find(&userChats).Error
	if err != nil {
		return nil, err
	}

	if len(userChats) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	userIDs := make([]string, len(userChats))
	for i, uc := range userChats {
		userIDs[i] = uc.UserID.String()
	}
	return userIDs, nil
}
