package repository

import (
	"log/slog"

	"example.com/messaging/src/internal/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type MessageRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *MessageRepository {
	return &MessageRepository{
		DB: db,
	}
}

func (r *MessageRepository) SaveMessage(message *models.Message) error {
	err := r.DB.Begin().Create(message).Commit().Error
	if err != nil {
		slog.Error("MessageRepositorySaveMessage failed: " + err.Error())
		return err
	}
	return nil
}

func (r *MessageRepository) GetDirectMessages(userId uuid.UUID, destinationId uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("user_id = ? AND destination_id = ? OR destination_id = ? AND user_id = ?", userId, destinationId, userId, destinationId).Find(&messages).Error
	if err != nil {
		slog.Error("MessageRepositoryGetDirectMessages failed: " + err.Error())
		return nil, err
	}
	return messages, nil
}

func (r *MessageRepository) GetMessagesByDestinationId(destinationId uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("destination_id = ?", destinationId).Find(&messages).Error
	if err != nil {
		slog.Error("MessageRepositoryGetMessagesByDestinationId failed: " + err.Error())
		return nil, err
	}
	return messages, nil
}
