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

func (r *ChannelRepository) FindById(channelID uuid.UUID) (*models.Channel, error) {
	var channel models.Channel
	err := r.db.Where("id = ?", channelID).First(&channel).Error
	return &channel, err
}

func (r *ChannelRepository) SaveChannel(channel dto.CreateChannelDTO) (uuid.UUID, error) {
	channelModel := models.Channel{
		Id:          uuid.New(),
		Name:        channel.Name,
		Description: channel.Description,
		CreatorId:   channel.CreatorId.String(),
	}
	err := r.db.Create(&channelModel).Error
	if err != nil {
		return uuid.Nil, err
	}
	return channelModel.Id, nil
}

func (r *ChannelRepository) DeleteChannel(channelID uuid.UUID) error {
	return r.db.Where("id = ?", channelID).Delete(&models.Channel{}).Error
}

func (r *ChannelRepository) UpdateChannel(channel dto.UpdateChannelDTO) error {
	return r.db.Model(&models.Channel{}).Where("id = ?", channel.Id).Updates(map[string]interface{}{
		"name":        channel.Name,
		"description": channel.Description,
	}).Error
}

func (r *ChannelRepository) AddUserToChannel(user dto.AddUserDTO) error {
	userChannel := models.UserChannel{
		ChannelId: user.ChannelId,
		UserId:    user.UserId,
	}
	return r.db.Create(&userChannel).Error
}

func (r *ChannelRepository) RemoveUserFromChannel(user dto.RemoveUserDTO) error {
	return r.db.Where("channel_id = ? AND user_id = ?", user.ChannelId, user.UserId).Delete(&models.UserChannel{}).Error
}

func (r *ChannelRepository) AddAdmin(admin dto.AdminDTO) error {
	adminModel := models.Admin{
		ChannelId: admin.ChannelId,
		UserId:    admin.UserId,
	}
	return r.db.Create(&adminModel).Error
}

func (r *ChannelRepository) RemoveAdmin(admin dto.AdminDTO) error {
	return r.db.Where("channel_id = ? AND user_id = ?", admin.ChannelId, admin.UserId).Delete(&models.Admin{}).Error
}

func (r *ChannelRepository) IsAdmin(channelID, userID uuid.UUID) (bool, error) {
	var admin models.Admin
	err := r.db.Where("channel_id = ? AND user_id = ?", channelID, userID).First(&admin).Error
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

func (r *ChannelRepository) GetChannelAdmins(channelID uuid.UUID) ([]string, error) {
	var admins []models.Admin
	err := r.db.Where("channel_id = ?", channelID).Find(&admins).Error
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
