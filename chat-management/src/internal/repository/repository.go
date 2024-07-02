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

func (r *ChatRepository) SaveChat(chat *models.Chat) error {
	return r.db.Create(chat).Error
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

func (r *ChatRepository) AddUserToChat(user *models.UserChat) error {
	return r.db.Create(&user).Error
}

func (r *ChatRepository) RemoveUserFromChat(user *models.UserChat) error {
	return r.db.Where("chat_id = ? AND user_id = ?", user.ChatId, user.UserId).Delete(&models.UserChat{}).Error
}

func (r *ChatRepository) AddAdmin(admin *models.Admin) error {
	return r.db.Create(&admin).Error
}

func (r *ChatRepository) RemoveAdmin(admin *models.Admin) error {
	return r.db.Where("chat_id = ? AND user_id = ?", admin.ChatId, admin.UserId).Delete(&models.Admin{}).Error
}

func (r *ChatRepository) IsAdmin(admin *models.Admin) (bool, error) {
	err := r.db.Where("chat_id = ? AND user_id = ?", admin.ChatId, admin.UserId).First(&admin).Error
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
