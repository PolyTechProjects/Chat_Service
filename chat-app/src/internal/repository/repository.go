package repository

import (
	"context"
	"log/slog"

	"example.com/chat-app/src/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type MessageRepository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func New(db *gorm.DB, redis *redis.Client) *MessageRepository {
	return &MessageRepository{
		DB:    db,
		Redis: redis,
	}
}

func (r *MessageRepository) SaveUserMessage(message *models.Message) error {
	if r.DB == nil {
		slog.Error("Database is not initialized")
	}
	err := r.DB.Begin().Create(message).Commit().Error
	return err
}

func (r *MessageRepository) GetMessageByChatRoomId(chatRoomId uuid.UUID) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("chat_room_id = ?", chatRoomId).Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) SubscribeToRedisChannel(channelName string) *redis.PubSub {
	return r.Redis.Subscribe(context.Background(), channelName)
}

func (r *MessageRepository) PublishToRedisChannel(channelName string, message interface{}) error {
	return r.Redis.Publish(context.Background(), channelName, message).Err()
}
