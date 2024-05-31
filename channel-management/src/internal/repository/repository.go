package repository

import (
	"example.com/channel-management/src/internal/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) SaveChannel(channel models.Channel) error {
	return r.db.Create(&channel).Error
}

func (r *ChannelRepository) DeleteChannel(channelID uuid.UUID) error {
	return r.db.Where("id = ?", channelID).Delete(&models.Channel{}).Error
}

func (r *ChannelRepository) UpdateChannel(channelID uuid.UUID, name, description string) error {
	return r.db.Model(&models.Channel{}).Where("id = ?", channelID).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error
}

func (r *ChannelRepository) AddUserToChannel(channelID, userID uuid.UUID) error {
	userChannel := models.UserChannel{
		ChannelID: channelID,
		UserID:    userID,
	}
	return r.db.Create(&userChannel).Error
}

func (r *ChannelRepository) RemoveUserFromChannel(channelID, userID uuid.UUID) error {
	return r.db.Where("channel_id = ? AND user_id = ?", channelID, userID).Delete(&models.UserChannel{}).Error
}

func (r *ChannelRepository) AddAdmin(channelID, userID uuid.UUID) error {
	admin := models.Admin{
		ChannelID: channelID,
		UserID:    userID,
	}
	return r.db.Create(&admin).Error
}

func (r *ChannelRepository) RemoveAdmin(channelID, userID uuid.UUID) error {
	return r.db.Where("channel_id = ? AND user_id = ?", channelID, userID).Delete(&models.Admin{}).Error
}

func (r *ChannelRepository) IsAdmin(channelID, userID uuid.UUID) (bool, error) {
	var admin models.Admin
	err := r.db.Where("channel_id = ? AND user_id = ?", channelID, userID).First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *ChannelRepository) GetChanUsers(channelId uuid.UUID) ([]string, error) {
	var userChannels []models.UserChannel
	err := r.db.Where("channel_id = ?", channelId).Find(&userChannels).Error
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, len(userChannels))
	for i, uc := range userChannels {
		userIDs[i] = uc.UserID.String()
	}
	return userIDs, nil
}
