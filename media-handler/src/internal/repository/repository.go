package repository

import (
	"example.com/media/src/internal/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type MediaHandlerRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *MediaHandlerRepository {
	return &MediaHandlerRepository{db: db}
}

func (m *MediaHandlerRepository) Save(media *models.Media) error {
	err := m.db.Create(&media).Error
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaHandlerRepository) FindById(id uuid.UUID) (*models.Media, error) {
	media := &models.Media{}
	err := m.db.Debug().Where("id = ?", id).Find(media).Error
	if err != nil {
		return nil, err
	}
	return media, nil
}

func (m *MediaHandlerRepository) DeleteById(id uuid.UUID) error {
	media := &models.Media{}
	err := m.db.Debug().Where("id = ?", id).Delete(media).Error
	if err != nil {
		return err
	}
	return nil
}
