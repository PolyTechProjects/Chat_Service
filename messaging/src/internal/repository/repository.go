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

func (r *MessageRepository) CreateMessage(message *models.Message) error {
	err := r.DB.Begin().Create(message).Commit().Error
	if err != nil {
		slog.Error("MessageRepositoryCreateMessage failed: " + err.Error())
		return err
	}
	return nil
}

func (r *MessageRepository) UpdateMessage(message *models.Message) error {
	err := r.DB.Begin().Save(message).Commit().Error
	if err != nil {
		slog.Error("MessageRepositoryUpdateMessage failed: " + err.Error())
		return err
	}
	return nil
}

func (r *MessageRepository) GetDirectMessages(userId uuid.UUID, destinationId uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("destination_id = ? AND is_deleted=false", destinationId).Order("created_at").Find(&messages).Error
	if err != nil {
		slog.Error("MessageRepositoryGetDirectMessages failed: " + err.Error())
		return nil, err
	}
	return messages, nil
}

func (r *MessageRepository) GetMessagesByDestinationId(destinationId uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("destination_id = ? AND is_deleted=false", destinationId).Order("created_at").Find(&messages).Error
	if err != nil {
		slog.Error("MessageRepositoryGetMessagesByDestinationId failed: " + err.Error())
		return nil, err
	}
	return messages, nil
}

func (r *MessageRepository) FindById(id uuid.UUID) (*models.Message, error) {
	message := &models.Message{}
	err := r.DB.Where("id = ?", id).First(message).Error
	if err != nil {
		slog.Error("MessageRepositoryFindById failed: " + err.Error())
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) DeleteMessage(message *models.Message) error {
	err := r.DB.Delete(message).Error
	if err != nil {
		slog.Error("MessageRepositoryDeleteMessage failed: " + err.Error())
		return err
	}
	return nil
}
