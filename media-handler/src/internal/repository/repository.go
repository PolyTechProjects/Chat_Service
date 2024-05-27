package repository

import (
	"context"
	"fmt"

	"example.com/media-handler/src/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type MediaHandlerRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func New(db *gorm.DB, redis *redis.Client) *MediaHandlerRepository {
	return &MediaHandlerRepository{db: db, redis: redis}
}

func (m *MediaHandlerRepository) PublishInFileLoadedChannel(message interface{}) error {
	return m.redis.Publish(context.Background(), "file-loaded-channel", message).Err()
}

func (m *MediaHandlerRepository) CacheVolumeIp(volumeId string, volumeIp string) error {
	_, err := m.redis.Set(context.Background(), fmt.Sprintf("VOLUME_%s", volumeId), volumeIp, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaHandlerRepository) GetVolumeIp(volumeId string) (string, error) {
	volumeIp, err := m.redis.Get(context.Background(), fmt.Sprintf("VOLUME_%s", volumeId)).Result()
	if err != nil {
		return "", err
	}
	return volumeIp, nil
}

func (m *MediaHandlerRepository) Save(media *models.Media) error {
	err := m.db.Create(&media).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaHandlerRepository) FindById(id uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := m.db.Debug().Where("id = ?", id).Find(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (m *MediaHandlerRepository) DeleteById(id uuid.UUID) error {
	err := m.db.Debug().Where("id = ?", id).Delete(&models.Media{}).Error
	if err != nil {
		return err
	}
	return nil
}
