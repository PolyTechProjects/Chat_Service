package repository

import (
	"log/slog"

	"example.com/chat/src/internal/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) FindById(chatId uuid.UUID) (*models.Chat, error) {
	chat := &models.Chat{}
	err := r.db.Where("id = ?", chatId).First(chat).Error
	if err != nil {
		slog.Error("ChatRepositoryFindById failed: " + err.Error())
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepository) FindByJoinLink(joinLink string) (*models.Chat, error) {
	chat := &models.Chat{}
	err := r.db.Where("join_link = ?", joinLink).First(&chat).Error
	if err != nil {
		slog.Error("ChatRepositoryFindByJoinLink failed: " + err.Error())
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepository) SaveChat(chat *models.Chat) error {
	err := r.db.Create(chat).Error
	if err != nil {
		slog.Error("ChatRepositorySaveChat failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatRepository) DeleteChat(chatId uuid.UUID) error {
	err := r.db.Where("id = ?", chatId).Delete(&models.Chat{}).Error
	if err != nil {
		slog.Error("ChatRepositoryDeleteChat failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatRepository) UpdateChat(chat *models.Chat) error {
	err := r.db.Save(chat).Error
	if err != nil {
		slog.Error("ChatRepositoryUpdateChat failed: " + err.Error())
		return err
	}
	return nil
}

type ChatUserRepository struct {
	db *gorm.DB
}

func NewChatUserRepository(db *gorm.DB) *ChatUserRepository {
	return &ChatUserRepository{db: db}
}

func (r *ChatUserRepository) FindByChat(chatId uuid.UUID) ([]*models.ChatUser, error) {
	var chatUsers []*models.ChatUser
	err := r.db.Where("chat_id = ?", chatId).Find(&chatUsers).Error
	if err != nil {
		slog.Error("ChatUserRepositoryFindByChat failed: " + err.Error())
		return nil, err
	}
	return chatUsers, nil
}

func (r *ChatUserRepository) FindByUser(userId uuid.UUID) ([]*models.ChatUser, error) {
	var chatUsers []*models.ChatUser
	err := r.db.Where("user_id = ?", userId).Find(&chatUsers).Error
	if err != nil {
		slog.Error("ChatUserRepositoryFindByUser failed: " + err.Error())
		return nil, err
	}
	return chatUsers, nil
}

func (r *ChatUserRepository) FindByChatAndUser(chatId uuid.UUID, userId uuid.UUID) (*models.ChatUser, error) {
	chatUser := &models.ChatUser{}
	err := r.db.Where("chat_id = ? AND user_id = ?", chatId, userId).First(chatUser).Error
	if err != nil {
		slog.Error("ChatUserRepositoryFindByChatAndUser failed: "+err.Error(), "requested chatId", chatId, "requested userId", userId)
		return nil, err
	}
	return chatUser, nil
}

func (r *ChatUserRepository) FindByChatAndRole(chatId uuid.UUID, roleId uuid.UUID) ([]*models.ChatUser, error) {
	var chatUsers []*models.ChatUser
	err := r.db.Where("chat_id = ? AND role_id = ?", chatId, roleId).Find(&chatUsers).Error
	if err != nil {
		slog.Error("ChatUserRepositoryFindByChatAndRole failed: " + err.Error())
		return nil, err
	}
	return chatUsers, nil
}

func (r *ChatUserRepository) AddChatUser(chatUser *models.ChatUser) error {
	err := r.db.Create(chatUser).Error
	if err != nil {
		slog.Error("ChatUserRepositoryAddChatUser failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatUserRepository) AddChatUsers(chatUsers []*models.ChatUser) error {
	for _, chatUser := range chatUsers {
		slog.Debug("User", "ptr", chatUser)
	}
	err := r.db.Create(chatUsers).Error
	if err != nil {
		slog.Error("ChatUserRepositoryAddChatUsers failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatUserRepository) DeleteChatUser(chatUser *models.ChatUser) error {
	err := r.db.Where("chat_id = ? AND user_id = ?", chatUser.ChatId, chatUser.UserId).Delete(&models.ChatUser{}).Error
	if err != nil {
		slog.Error("ChatUserRepositoryDeleteChatUser failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatUserRepository) DeleteChatUsers(chatUsers []*models.ChatUser) error {
	for _, chatUser := range chatUsers {
		err := r.db.Where("chat_id = ? AND user_id = ?", chatUser.ChatId, chatUser.UserId).Delete(&models.ChatUser{}).Error
		if err != nil {
			slog.Error("ChatUserRepositoryDeleteChatUsers failed: " + err.Error())
			return err
		}
	}
	return nil
}

func (r *ChatUserRepository) DeleteByChat(chatId uuid.UUID) error {
	err := r.db.Where("chat_id = ?", chatId).Delete(&models.ChatUser{}).Error
	if err != nil {
		slog.Error("ChatUserRepositoryDeleteByChat failed: " + err.Error())
		return err
	}
	return nil
}

func (r *ChatUserRepository) UpdateChatUser(chatUser *models.ChatUser) error {
	err := r.db.Save(chatUser).Error
	if err != nil {
		slog.Error("ChatUserRepositoryUpdateChatUser failed: " + err.Error())
		return err
	}
	return nil
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindById(roleId uuid.UUID) (*models.Role, error) {
	role := &models.Role{}
	err := r.db.Where("id = ?", roleId).First(role).Error
	if err != nil {
		slog.Error("RoleRepositoryFindById failed: " + err.Error())
		return nil, err
	}
	return role, err
}

func (r *RoleRepository) AddRole(role *models.Role) error {
	err := r.db.Create(role).Error
	if err != nil {
		slog.Error("RoleRepositoryAddRole failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RoleRepository) UpdateRole(role *models.Role) error {
	err := r.db.Save(role).Error
	if err != nil {
		slog.Error("RoleRepositoryUpdateRole failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RoleRepository) DeleteRole(roleId uuid.UUID) error {
	err := r.db.Where("id = ?", roleId).Delete(&models.Role{}).Error
	if err != nil {
		slog.Error("RoleRepositoryDeleteRole failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RoleRepository) GetRoles() ([]*models.Role, error) {
	var roles []*models.Role
	err := r.db.Find(&roles).Error
	if err != nil {
		slog.Error("RoleRepositoryGetRoles failed: " + err.Error())
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) FindByChat(chatId uuid.UUID) ([]*models.Role, error) {
	var chatRoles []*models.Role
	err := r.db.Where("chat_id = ?", chatId).Find(&chatRoles).Error
	if err != nil {
		slog.Error("RoleRepositoryFindByChat failed: " + err.Error())
		return nil, err
	}
	return chatRoles, nil
}

func (r *RoleRepository) FindDefaultRole(chatId uuid.UUID) (*models.Role, error) {
	chatRole := &models.Role{}
	err := r.db.Where("chat_id = ? AND is_default = ?", chatId, true).Find(&chatRole).Error
	if err != nil {
		slog.Error("RoleRepositoryFindDefaultRole failed: " + err.Error())
		return nil, err
	}
	return chatRole, nil
}

func (r *RoleRepository) DeleteByChat(chatId uuid.UUID) error {
	err := r.db.Where("chat_id = ?", chatId).Delete(&models.Role{}).Error
	if err != nil {
		slog.Error("RoleRepositoryDeleteByChat failed: " + err.Error())
		return err
	}
	return nil
}

type RolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

func (r *RolePermissionRepository) FindByRole(roleId uuid.UUID) ([]*models.RolePermission, error) {
	var rolePermissions []*models.RolePermission
	err := r.db.Where("role_id = ?", roleId).Find(&rolePermissions).Error
	if err != nil {
		slog.Error("RolePermissionRepositoryFindByRole failed: " + err.Error())
		return nil, err
	}
	return rolePermissions, nil
}

func (r *RolePermissionRepository) AddRolePermission(rolePermission *models.RolePermission) error {
	err := r.db.Create(rolePermission).Error
	if err != nil {
		slog.Error("RolePermissionRepositoryAddRolePermission failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RolePermissionRepository) AddRolePermissions(rolePermissions []*models.RolePermission) error {
	if rolePermissions == nil {
		slog.Info("NIL")
	}
	err := r.db.Save(rolePermissions).Error
	if err != nil {
		slog.Error("RolePermissionRepositoryAddRolePermissions failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RolePermissionRepository) DeleteRolePermission(rolePermission *models.RolePermission) error {
	err := r.db.Where("role_id = ? AND permission = ?", rolePermission.RoleId, rolePermission.Permission).Delete(&models.RolePermission{}).Error
	if err != nil {
		slog.Error("RolePermissionRepositoryDeleteRolePermission failed: " + err.Error())
		return err
	}
	return nil
}

func (r *RolePermissionRepository) DeleteRolePermissions(rolePermissions []*models.RolePermission) error {
	for _, rolePermission := range rolePermissions {
		err := r.db.Where("role_id = ? AND permission = ?", rolePermission.RoleId, rolePermission.Permission).Delete(&models.RolePermission{}).Error
		if err != nil {
			slog.Error("RepositoryDeleteRolePermissions failed: " + err.Error())
			return err
		}
	}
	return nil
}

func (r *RolePermissionRepository) DeleteByRole(roleId uuid.UUID) error {
	err := r.db.Where("role_id = ?", roleId).Delete(&models.RolePermission{}).Error
	if err != nil {
		slog.Error("RolePermissionRepositoryDeleteByRole failed: " + err.Error())
		return err
	}
	return nil
}

type DirectChatRepository struct {
	db *gorm.DB
}

func NewDirectChatRepository(db *gorm.DB) *DirectChatRepository {
	return &DirectChatRepository{db: db}
}

func (r *DirectChatRepository) FindByChat(chatId uuid.UUID) (*models.DirectChat, error) {
	directChat := &models.DirectChat{}
	err := r.db.Where("id = ?", chatId).Find(directChat).Error
	if err != nil {
		slog.Error("DirectChatRepositoryFindByChat failed: " + err.Error())
		return nil, err
	}
	return directChat, nil
}

func (r *DirectChatRepository) AddDirectChat(directChat *models.DirectChat) error {
	err := r.db.Create(directChat).Error
	if err != nil {
		slog.Error("DirectChatRepositoryAddDirectChat failed: " + err.Error())
		return err
	}
	return nil
}

func (r *DirectChatRepository) DeleteDirectChat(directChat *models.DirectChat) error {
	err := r.db.Where("id = ?", directChat.Id).Delete(&models.DirectChat{}).Error
	if err != nil {
		slog.Error("DirectChatRepositoryDeleteDirectChat failed: " + err.Error())
		return err
	}
	return nil
}

func (r *DirectChatRepository) DeleteDirectChats(directChats []*models.DirectChat) error {
	r.db.Begin()
	for _, directChat := range directChats {
		err := r.db.Where("id = ?", directChat.Id).Delete(&models.DirectChat{}).Error
		if err != nil {
			slog.Error("RepositoryDeleteDirectChats failed: " + err.Error())
			r.db.Rollback()
			return err
		}
	}
	r.db.Commit()
	return nil
}

func (r *DirectChatRepository) FindByUser(userId uuid.UUID) ([]*models.DirectChat, error) {
	var directChats []*models.DirectChat
	err := r.db.Where("first_user_id = ? OR second_user_id = ?", userId, userId).Find(&directChats).Error
	if err != nil {
		slog.Error("DirectChatRepositoryFindByUser failed: " + err.Error())
		return nil, err
	}
	return directChats, nil
}

func (r *DirectChatRepository) FindByUsers(firstUserId uuid.UUID, secondUserId uuid.UUID) (*models.DirectChat, error) {
	directChat := &models.DirectChat{}
	err := r.db.Where("first_user_id = ? AND second_user_id = ?", firstUserId, secondUserId).Find(&directChat).Error
	if err != nil {
		slog.Error("DirectChatRepositoryFindByUsers failed: " + err.Error())
		return nil, err
	}
	return directChat, nil
}
