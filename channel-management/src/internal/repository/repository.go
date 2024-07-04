package repository

import (
	"example.com/channel-management/src/internal/dto"
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

func (r *ChannelRepository) FindById(channelId uuid.UUID) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.Where("id = ?", channelId).First(&channel).Error
	return &channel, err
}

func (r *ChannelRepository) SaveChannel(channel *models.Channel) (uuid.UUID, error) {
	err := r.db.Create(&channel).Error
	if err != nil {
		return uuid.Nil, err
	}
	return channel.Id, nil
}

func (r *ChannelRepository) DeleteChannel(channelID uuid.UUID) error {
	return r.db.Where("id = ?", channelID).Delete(&models.Channel{}).Error
}

func (r *ChannelRepository) UpdateChannel(channel *dto.UpdateChannelRequest) error {
	return r.db.Model(&models.Channel{}).Where("id = ?", channel.ChannelId).Updates(map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
	}).Error
}

func (r *ChannelRepository) AddUserToChannel(user *models.UserChannel) error {
	return r.db.Create(user).Error
}

func (r *ChannelRepository) RemoveUserFromChannel(user *models.UserChannel) error {
	return r.db.Where("channel_id = ? AND user_id = ?", user.ChannelId, user.UserId).Delete(&models.UserChannel{}).Error
}

func (r *ChannelRepository) AddAdmin(admin *models.Admin) error {
	return r.db.Create(&admin).Error
}

func (r *ChannelRepository) RemoveAdmin(admin *models.Admin) error {
	return r.db.Where("channel_id = ? AND user_id = ?", admin.ChannelId, admin.UserId).Delete(&models.Admin{}).Error
}

func (r *ChannelRepository) IsAdmin(admin *models.Admin) (bool, error) {
	err := r.db.Where("channel_id = ? AND user_id = ?", admin.ChannelId, admin.UserId).First(&admin).Error
	if err != nil {
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

	if len(userChannels) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	userIDs := make([]string, len(userChannels))
	for i, uc := range userChannels {
		userIDs[i] = uc.UserId.String()
	}
	return userIDs, nil
}

func (r *ChannelRepository) GetChannelAdmins(channelId uuid.UUID) ([]string, error) {
	var admins []models.Admin
	err := r.db.Where("channel_id = ?", channelId).Find(&admins).Error
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
