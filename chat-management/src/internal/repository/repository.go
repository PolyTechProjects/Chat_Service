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

func (r *ChatRepository) FindById(chatId uuid.UUID) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Where("id = ?", chatId).First(&chat).Error
	return &chat, err
}

func (r *ChatRepository) SaveChat(chat models.Chat) error {
	return r.db.Create(&chat).Error
}

func (r *ChatRepository) DeleteChat(chatId uuid.UUID) error {
	return r.db.Where("id = ?", chatId).Delete(&models.Chat{}).Error
}

func (r *ChatRepository) UpdateChat(request *dto.UpdateChatRequest) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", request.ChatId).Updates(map[string]interface{}{
		"name":        request.Name,
		"description": request.Description,
	}).Error
}

func (r *ChatRepository) AddUserToChat(request *dto.UserChatRequest) error {
	userChat := models.UserChat{
		ChatId: request.ChatId,
		UserId: request.UserId,
	}
	return r.db.Create(&userChat).Error
}

func (r *ChatRepository) RemoveUserFromChat(request *dto.UserChatRequest) error {
	return r.db.Where("chat_id = ? AND user_id = ?", request.ChatId, request.UserId).Delete(&models.UserChat{}).Error
}

func (r *ChatRepository) AddAdmin(request *dto.AdminRequest) error {
	admin := models.Admin{
		ChatId: request.ChatId,
		UserId: request.UserId,
	}
	return r.db.Create(&admin).Error
}

func (r *ChatRepository) RemoveAdmin(request *dto.AdminRequest) error {
	return r.db.Where("chat_id = ? AND user_id = ?", request.ChatId, request.UserId).Delete(&models.Admin{}).Error
}

func (r *ChatRepository) IsAdmin(request *dto.AdminRequest) (bool, error) {
	var admin models.Admin
	err := r.db.Where("chat_id = ? AND user_id = ?", request.ChatId, request.UserId).First(&admin).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ChatRepository) IsMember(request *dto.UserChatRequest) (bool, error) {
	var userChat models.UserChat
	err := r.db.Where("chat_id = ? AND user_id = ?", request.ChatId, request.UserId).First(&userChat).Error
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
		userIDs[i] = uc.UserId.String()
	}
	return userIDs, nil
}

func (r *ChatRepository) GetChatAdmins(chatId uuid.UUID) ([]string, error) {
	var admins []models.Admin
	err := r.db.Where("chat_id = ?", chatId).Find(&admins).Error
	if err != nil {
		return nil, err
	}

	if len(admins) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	userIds := make([]string, len(admins))
	for i, a := range admins {
		userIds[i] = a.UserId.String()
	}
	return userIds, nil
}
