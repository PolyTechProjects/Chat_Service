package repository

import (
	"log/slog"

	"example.com/chat-app/src/internal/models"
	"github.com/jinzhu/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) SaveUserMessage(message *models.Message) error {
	if r.DB == nil {
		slog.Error("Database is not initialized")
	}
	err := r.DB.Begin().Create(*message).Commit().Error
	return err
}

func (r *Repository) GetMessageByChatRoomId(chatRoomId uint64) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("chat_room_id = ?", chatRoomId).Find(&messages).Error
	return messages, err
}
