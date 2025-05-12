package repository

import (
	"log/slog"

	"example.com/notification/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) AddSubscription(subscription *models.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *SubscriptionRepository) DeleteSubscriptions(subscriptions []*models.Subscription) error {
	r.db.Begin()
	for _, subscription := range subscriptions {
		err := r.db.Where("chat_id = ? AND user_id = ?", subscription.ChatId, subscription.UserId).Delete(&models.Subscription{}).Error
		if err != nil {
			slog.Error("RepositoryDeleteSubscriptions failed: " + err.Error())
			r.db.Rollback()
			return err
		}
	}
	r.db.Commit()
	return nil
}

func (r *SubscriptionRepository) DeleteSubscriptionsByChatId(chatId uuid.UUID) error {
	err := r.db.Where("chat_id = ?", chatId).Delete(&models.Subscription{}).Error
	if err != nil {
		slog.Error("RepositoryDeleteSubscriptionsByChatId failed: " + err.Error())
		return err
	}
	return nil
}

func (r *SubscriptionRepository) DeleteSubscriptionsByUserId(userId uuid.UUID) error {
	err := r.db.Where("user_id = ?", userId).Delete(&models.Subscription{}).Error
	if err != nil {
		slog.Error("RepositoryDeleteSubscriptionsByUserId failed: " + err.Error())
		return err
	}
	return nil
}

func (r *SubscriptionRepository) DeleteSubscriptionsByChatIdAndUserId(chatId uuid.UUID, userId uuid.UUID) error {
	err := r.db.Where("chat_id = ? AND user_id = ?", chatId, userId).Delete(&models.Subscription{}).Error
	if err != nil {
		slog.Error("RepositoryDeleteSubscriptionsByChatIdAndUserId failed: " + err.Error())
		return err
	}
	return nil
}

func (r *SubscriptionRepository) FindSubscriptionsByChatId(chatId uuid.UUID) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.Where("chat_id = ?", chatId).Find(&subscriptions).Error
	if err != nil {
		slog.Error("RepositoryFindSubscriptionsByChatId failed: " + err.Error())
		return nil, err
	}
	return subscriptions, nil
}

func (r *SubscriptionRepository) FindSubscriptionsByUserId(userId uuid.UUID) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := r.db.Where("user_id = ?", userId).Find(&subscriptions).Error
	if err != nil {
		slog.Error("RepositoryFindSubscriptionsByUserId failed: " + err.Error())
		return nil, err
	}
	return subscriptions, nil
}

func (r *SubscriptionRepository) FindSubscriptionsByChatIdAndUserId(chatId uuid.UUID, userId uuid.UUID) (*models.Subscription, error) {
	subscription := &models.Subscription{}
	err := r.db.Where("chat_id = ? AND user_id = ?", chatId, userId).First(subscription).Error
	if err != nil {
		slog.Error("RepositoryFindSubscriptionsByChatIdAndUserId failed: " + err.Error())
		return nil, err
	}
	return subscription, nil
}

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) AddNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) FindById(notificationId uuid.UUID) (*models.Notification, error) {
	notification := &models.Notification{}
	err := r.db.Where("id = ?", notificationId).First(notification).Error
	if err != nil {
		slog.Error("NotificationRepositoryFindById failed: " + err.Error())
		return nil, err
	}
	return notification, nil
}
