package service

import (
	"fmt"

	"example.com/chat/src/internal/client"
	"example.com/chat/src/internal/dto"
	"example.com/chat/src/internal/models"
	"example.com/chat/src/internal/repository"
	"github.com/google/uuid"
)

type ChatService struct {
	ChatRepository           *repository.ChatRepository
	DirectChatRepository     *repository.DirectChatRepository
	ChatUserRepository       *repository.ChatUserRepository
	RoleRepository           *repository.RoleRepository
	RolePermissionRepository *repository.RolePermissionRepository
	ChatRoleRepository       *repository.ChatRoleRepository
	UsersClient              *client.UsersGRPCClient
}

func NewChatService(
	chatRepository *repository.ChatRepository,
	directChatRepository *repository.DirectChatRepository,
	chatUserRepository *repository.ChatUserRepository,
	roleRepository *repository.RoleRepository,
	rolePermissionRepository *repository.RolePermissionRepository,
	chatRoleRepository *repository.ChatRoleRepository,
	usersClient *client.UsersGRPCClient,
) *ChatService {
	return &ChatService{
		ChatRepository:           chatRepository,
		DirectChatRepository:     directChatRepository,
		ChatUserRepository:       chatUserRepository,
		RoleRepository:           roleRepository,
		RolePermissionRepository: rolePermissionRepository,
		ChatRoleRepository:       chatRoleRepository,
		UsersClient:              usersClient,
	}
}

func (s *ChatService) GetChat(chatId uuid.UUID, userId uuid.UUID) (*dto.GetChatResponse, error) {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return nil, err
	}
	if !userInChat {
		return nil, fmt.Errorf("user is not in chat")
	}
	chat, err := s.ChatRepository.FindById(chatId)
	if err != nil {
		return nil, err
	}
	chatUsers, err := s.ChatUserRepository.FindByChat(chatId)
	if err != nil {
		return nil, err
	}
	participants := make([]*dto.Participants, len(chatUsers))
	for i, chatUser := range chatUsers {
		participants[i] = &dto.Participants{
			UserId:   chatUser.UserId.String(),
			RoleId:   chatUser.RoleId.String(),
			Nickname: chatUser.Nickname,
		}
	}
	getResp := &dto.GetChatResponse{
		Chat:         chat,
		Participants: participants,
	}
	return getResp, nil
}

func (s *ChatService) DeleteChat(chatId uuid.UUID, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doDeleteChat(chatId)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_DELETE_CHAT)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doDeleteChat(chatId)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) CreateChat(req *dto.CreateChatRequest) (*dto.GetChatResponse, error) {
	chat := &models.Chat{
		Id:          uuid.New(),
		Name:        req.Name,
		CreatorId:   req.CreatorId,
		IsChannel:   req.IsChannel,
		IsClosed:    req.IsClosed,
		JoinLink:    req.JoinLink,
		ProfilePic:  req.ProfilePic,
		Description: req.Description,
	}
	err := s.ChatRepository.SaveChat(chat)
	if err != nil {
		return nil, err
	}

	defaultRole := &models.Role{
		Id:          uuid.New(),
		Name:        "default",
		Description: "Default role",
		Color:       "#000000",
	}
	err = s.RoleRepository.AddRole(defaultRole)
	if err != nil {
		return nil, err
	}
	chatRole := &models.ChatRole{
		ChatId:    chat.Id,
		RoleId:    defaultRole.Id,
		IsDefault: true,
	}
	err = s.ChatRoleRepository.AddChatRole(chatRole)
	if err != nil {
		return nil, err
	}
	rolePermissions := make([]*models.RolePermission, len(req.DefaultPermissions))
	for i, permission := range req.DefaultPermissions {
		rolePermission := &models.RolePermission{
			RoleId:     defaultRole.Id,
			Permission: string(permission),
		}
		rolePermissions[i] = rolePermission
	}
	err = s.RolePermissionRepository.AddRolePermissions(rolePermissions)
	if err != nil {
		return nil, err
	}

	defaultAdminRole := &models.Role{
		Id:          uuid.New(),
		Name:        "admin",
		Description: "Admin role",
		Color:       "#ffffff",
	}
	err = s.RoleRepository.AddRole(defaultAdminRole)
	if err != nil {
		return nil, err
	}
	chatRole = &models.ChatRole{
		ChatId:  chat.Id,
		RoleId:  defaultAdminRole.Id,
		IsAdmin: true,
	}
	err = s.ChatRoleRepository.AddChatRole(chatRole)
	if err != nil {
		return nil, err
	}
	defaultAdminPermissions := getAllPermissions()
	adminRolePermissions := make([]*models.RolePermission, len(defaultAdminPermissions))
	for i, permission := range defaultAdminPermissions {
		rolePermission := &models.RolePermission{
			RoleId:     defaultAdminRole.Id,
			Permission: string(permission),
		}
		adminRolePermissions[i] = rolePermission
	}
	err = s.RolePermissionRepository.AddRolePermissions(adminRolePermissions)
	if err != nil {
		return nil, err
	}

	participants := make([]*dto.Participants, len(req.ParticipantsIds))
	userIds := append(req.ParticipantsIds, req.CreatorId)
	names, err := s.UsersClient.PerformGetUsers(userIds)
	if err != nil {
		return nil, err
	}
	users := make([]*models.ChatUser, len(userIds))
	for i, participantId := range req.ParticipantsIds {
		chatUser := &models.ChatUser{
			ChatId:   chat.Id,
			UserId:   participantId,
			RoleId:   defaultRole.Id,
			Nickname: names.Users[i].Name,
		}
		users[i] = chatUser
		participants[i] = &dto.Participants{
			UserId:   chatUser.UserId.String(),
			RoleId:   chatUser.RoleId.String(),
			Nickname: chatUser.Nickname,
		}
	}

	creator := &models.ChatUser{
		ChatId:   chat.Id,
		UserId:   req.CreatorId,
		RoleId:   defaultAdminRole.Id,
		Nickname: names.Users[len(userIds)-1].Name,
	}
	users[len(userIds)-1] = creator
	participants[len(userIds)-1] = &dto.Participants{
		UserId:   creator.UserId.String(),
		RoleId:   creator.RoleId.String(),
		Nickname: creator.Nickname,
	}
	err = s.ChatUserRepository.AddChatUsers(users)
	if err != nil {
		return nil, err
	}

	return &dto.GetChatResponse{Chat: chat, Participants: participants}, nil
}

func (s *ChatService) EditChat(req *dto.EditChatRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doEditChat(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_EDIT_CHAT)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doEditChat(req)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) JoinChat(joinLink string, userId uuid.UUID) (*dto.GetChatResponse, error) {
	chat, err := s.ChatRepository.FindByJoinLink(joinLink)
	if err != nil {
		return nil, err
	}
	defaultRole, err := s.ChatRoleRepository.FindDefaultRole(chat.Id)
	if err != nil {
		return nil, err
	}
	chatUser := &models.ChatUser{
		ChatId:   chat.Id,
		UserId:   userId,
		RoleId:   defaultRole.RoleId,
		Nickname: "Anonymous",
	}
	err = s.ChatUserRepository.AddChatUser(chatUser)
	if err != nil {
		return nil, err
	}
	return &dto.GetChatResponse{Chat: chat}, nil
}

func (s *ChatService) AddUsers(req *dto.AddUsersRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doAddUsers(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_ADD_USERS)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doAddUsers(req)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) DeleteUsers(chatId uuid.UUID, userIds []uuid.UUID, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doDeleteUsers(chatId, userIds)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_DELETE_USERS)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doDeleteUsers(chatId, userIds)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) ChangeUserNickname(req *dto.ChangeUserNicknameRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doChangeUserNickname(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	var requiredPermission string
	if req.UserId == userId {
		requiredPermission = string(models.CAN_CHANGE_OWN_NICKNAME)
	} else {
		requiredPermission = string(models.CAN_CHANGE_OTHERS_NICKNAME)
	}
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doChangeUserNickname(req)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) CreateRole(req *dto.CreateRoleRequest, userId uuid.UUID) (*dto.RoleResponse, error) {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	if !userInChat {
		return nil, fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return nil, err
	}
	if role.IsAdmin {
		return s.doCreateRole(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return nil, err
	}
	requiredPermission := string(models.CAN_CREATE_ROLE)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doCreateRole(req)
		}
	}
	return nil, fmt.Errorf("permission denied")
}

func (s *ChatService) EditRole(req *dto.UpdateRoleRequest, userId uuid.UUID) (*dto.RoleResponse, error) {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	if !userInChat {
		return nil, fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return nil, err
	}
	if role.IsAdmin {
		return s.doEditRole(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return nil, err
	}
	requiredPermission := string(models.CAN_EDIT_ROLE)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doEditRole(req)
		}
	}
	return nil, fmt.Errorf("permission denied")
}

func (s *ChatService) DeleteRole(chatId uuid.UUID, userId uuid.UUID, roleId uuid.UUID) error {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	targetRole, err := s.ChatRoleRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	if targetRole.IsDefault || targetRole.IsAdmin {
		return fmt.Errorf("cannot delete default/admin role")
	}
	role, err := s.ChatRoleRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doDeleteRole(chatId, roleId)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_DELETE_ROLE)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doDeleteRole(chatId, roleId)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) SetRole(req *dto.SetRoleRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, userId)
	if err != nil {
		return err
	}
	role, err := s.ChatRoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin {
		return s.doSetRole(req)
	}
	rolePermissions, err := s.RolePermissionRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	requiredPermission := string(models.CAN_SET_ROLES)
	for _, permission := range rolePermissions {
		if permission.Permission == requiredPermission {
			return s.doSetRole(req)
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) GetDirectChat(chatId uuid.UUID, userId uuid.UUID) (*dto.DirectChatResponse, error) {
	directChat, err := s.DirectChatRepository.FindByChat(chatId)
	if err != nil {
		return nil, err
	}
	if directChat.FirstUserId != userId && directChat.SecondUserId != userId {
		return nil, fmt.Errorf("user is not in chat")
	}
	return &dto.DirectChatResponse{
		ChatId:       directChat.Id,
		FirstUserId:  directChat.FirstUserId,
		SecondUserId: directChat.SecondUserId,
	}, nil
}

func (s *ChatService) CreateDirectChat(req *dto.CreateDirectChatRequest, userId uuid.UUID) (*dto.DirectChatResponse, error) {
	directChat := &models.DirectChat{
		FirstUserId:  req.FirstUserId,
		SecondUserId: req.SecondUserId,
	}
	err := s.DirectChatRepository.AddDirectChat(directChat)
	if err != nil {
		return nil, err
	}
	return &dto.DirectChatResponse{
		ChatId:       directChat.Id,
		FirstUserId:  directChat.FirstUserId,
		SecondUserId: directChat.SecondUserId,
	}, nil
}

func (s *ChatService) DeleteDirectChat(chatId uuid.UUID, userId uuid.UUID) error {
	directChat, err := s.DirectChatRepository.FindByChat(chatId)
	if err != nil {
		return err
	}
	if directChat.FirstUserId != userId && directChat.SecondUserId != userId {
		return fmt.Errorf("user is not in chat")
	}
	return s.DirectChatRepository.DeleteDirectChat(directChat)
}

func (s *ChatService) DeleteDirectChats(chatIds []uuid.UUID, userId uuid.UUID) error {
	directChats := make([]*models.DirectChat, len(chatIds))
	for i, chatId := range chatIds {
		directChat, err := s.DirectChatRepository.FindByChat(chatId)
		if err != nil {
			return err
		}
		if directChat.FirstUserId != userId && directChat.SecondUserId != userId {
			return fmt.Errorf("user is not in chat")
		}
		directChats[i] = directChat
	}
	err := s.DirectChatRepository.DeleteDirectChats(directChats)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) SearchDirectChat(firstUserId uuid.UUID, secondUserId uuid.UUID) (*dto.DirectChatResponse, error) {
	directChat, err := s.DirectChatRepository.FindByUsers(firstUserId, secondUserId)
	if err != nil {
		return nil, err
	}
	return &dto.DirectChatResponse{
		ChatId:       directChat.Id,
		FirstUserId:  directChat.FirstUserId,
		SecondUserId: directChat.SecondUserId,
	}, nil
}

func (s *ChatService) GetChats(userId uuid.UUID) (*dto.ChatsResponse, error) {
	chatUsers, err := s.ChatUserRepository.FindByUser(userId)
	if err != nil {
		return nil, err
	}
	directChats, err := s.DirectChatRepository.FindByUser(userId)
	if err != nil {
		return nil, err
	}
	responseChats := make([]*models.Chat, len(chatUsers))
	for i, chatUser := range chatUsers {
		chat, err := s.ChatRepository.FindById(chatUser.ChatId)
		if err != nil {
			return nil, err
		}
		responseChats[i] = chat
	}
	responseDirectChats := make([]*dto.DirectChatResponse, len(directChats))
	for i, directChat := range directChats {
		responseDirectChats[i] = &dto.DirectChatResponse{
			ChatId:       directChat.Id,
			FirstUserId:  directChat.FirstUserId,
			SecondUserId: directChat.SecondUserId,
		}
	}
	return &dto.ChatsResponse{Chats: responseChats, DirectChats: responseDirectChats}, nil
}

func (s *ChatService) VerifyUserAction(chatId uuid.UUID, userId uuid.UUID, action string) error {
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	roleId := chatUser.RoleId
	rolePermission, err := s.RolePermissionRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	preferablePermission := "CAN_" + action
	for _, permission := range rolePermission {
		if permission.Permission == preferablePermission {
			return nil
		}
	}
	return fmt.Errorf("permission denied")
}

func getAllPermissions() []models.Permission {
	return []models.Permission{
		models.CAN_WRITE_MESSAGE,
		models.CAN_PIN_MESSAGE,
		models.CAN_EDIT_OWN_MESSAGE,
		models.CAN_EDIT_OTHERS_MESSAGE,
		models.CAN_DELETE_OWN_MESSAGE,
		models.CAN_DELETE_OTHERS_MESSAGE,
		models.CAN_SEND_FILE,
		models.CAN_CHANGE_OWN_NICKNAME,
		models.CAN_CHANGE_OTHERS_NICKNAME,
		models.CAN_EDIT_CHAT,
		models.CAN_DELETE_CHAT,
		models.CAN_CREATE_ROLE,
		models.CAN_EDIT_ROLE,
		models.CAN_DELETE_ROLE,
		models.CAN_SET_ROLES,
		models.CAN_ADD_USERS,
		models.CAN_DELETE_USERS,
	}
}

func (s *ChatService) isUserInChat(chatId uuid.UUID, userId uuid.UUID) (bool, error) {
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return false, err
	}
	if chatUser == nil {
		return false, nil
	}
	return true, nil
}

func (s *ChatService) doDeleteChat(chatId uuid.UUID) error {
	err := s.ChatUserRepository.DeleteByChat(chatId)
	if err != nil {
		return err
	}
	chatRoles, err := s.ChatRoleRepository.FindByChat(chatId)
	if err != nil {
		return err
	}
	err = s.ChatRoleRepository.DeleteByChat(chatId)
	if err != nil {
		return err
	}
	for _, chatRole := range chatRoles {
		err = s.RolePermissionRepository.DeleteByRole(chatRole.RoleId)
		if err != nil {
			return err
		}
		err = s.RoleRepository.DeleteRole(chatRole.RoleId)
		if err != nil {
			return err
		}
	}
	err = s.ChatRepository.DeleteChat(chatId)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doEditChat(req *dto.EditChatRequest) error {
	chat, err := s.ChatRepository.FindById(req.ChatId)
	if err != nil {
		return err
	}
	chat.Name = req.Name
	chat.ProfilePic = req.ProfilePic
	chat.JoinLink = req.JoinLink
	chat.Description = req.Description
	err = s.ChatRepository.UpdateChat(chat)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doAddUsers(req *dto.AddUsersRequest) error {
	chatUsers := make([]*models.ChatUser, len(req.UserIds))
	defaultRole, err := s.ChatRoleRepository.FindDefaultRole(req.ChatId)
	if err != nil {
		return err
	}
	users, err := s.UsersClient.PerformGetUsers(req.UserIds)
	if err != nil {
		return err
	}
	for i, userId := range req.UserIds {
		chatUsers[i] = &models.ChatUser{
			ChatId:   req.ChatId,
			UserId:   userId,
			RoleId:   defaultRole.RoleId,
			Nickname: users.Users[i].Name,
		}
	}
	err = s.ChatUserRepository.AddChatUsers(chatUsers)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doDeleteUsers(chatId uuid.UUID, userIds []uuid.UUID) error {
	chat, err := s.ChatRepository.FindById(chatId)
	if err != nil {
		return err
	}
	chatUsers := make([]*models.ChatUser, len(userIds))
	for i, userId := range userIds {
		if chat.CreatorId == userId {
			return fmt.Errorf("cannot delete creator")
		}
		chatUsers[i] = &models.ChatUser{
			ChatId: chatId,
			UserId: userId,
		}
	}
	err = s.ChatUserRepository.DeleteChatUsers(chatUsers)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doChangeUserNickname(req *dto.ChangeUserNicknameRequest) error {
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, req.UserId)
	if err != nil {
		return err
	}
	chatUser.Nickname = req.Nickname
	err = s.ChatUserRepository.UpdateChatUser(chatUser)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doCreateRole(req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	newRole := &models.Role{
		Id:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}
	err := s.RoleRepository.AddRole(newRole)
	if err != nil {
		return nil, err
	}
	chatRole := &models.ChatRole{
		ChatId:    req.ChatId,
		RoleId:    newRole.Id,
		IsDefault: false,
	}
	err = s.ChatRoleRepository.AddChatRole(chatRole)
	if err != nil {
		return nil, err
	}
	newRolePermissions := make([]*models.RolePermission, len(req.Permissions))
	for i, permission := range req.Permissions {
		rolePermission := &models.RolePermission{
			Permission: permission,
			RoleId:     newRole.Id,
		}
		newRolePermissions[i] = rolePermission
	}
	err = s.RolePermissionRepository.AddRolePermissions(newRolePermissions)
	if err != nil {
		return nil, err
	}
	return &dto.RoleResponse{
		RoleId:      newRole.Id,
		ChatId:      req.ChatId,
		Name:        newRole.Name,
		Description: newRole.Description,
		Color:       newRole.Color,
		Permissions: req.Permissions,
	}, nil
}

func (s *ChatService) doEditRole(req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.RoleRepository.FindById(req.RoleId)
	if err != nil {
		return nil, err
	}
	role.Name = req.Name
	role.Description = req.Description
	role.Color = req.Color
	err = s.RoleRepository.UpdateRole(role)
	if err != nil {
		return nil, err
	}
	oldRolePermissions, err := s.RolePermissionRepository.FindByRole(req.RoleId)
	if err != nil {
		return nil, err
	}
	rolePermissions := make([]*models.RolePermission, len(req.Permissions))
	for i, permission := range req.Permissions {
		rolePermission := &models.RolePermission{
			Permission: permission,
			RoleId:     req.RoleId,
		}
		rolePermissions[i] = rolePermission
	}
	err = s.RolePermissionRepository.DeleteRolePermissions(oldRolePermissions)
	if err != nil {
		return nil, err
	}
	err = s.RolePermissionRepository.AddRolePermissions(rolePermissions)
	if err != nil {
		return nil, err
	}
	return &dto.RoleResponse{
		RoleId:      req.RoleId,
		ChatId:      req.ChatId,
		Name:        role.Name,
		Description: role.Description,
		Color:       role.Color,
		Permissions: req.Permissions,
	}, nil
}

func (s *ChatService) doDeleteRole(chatId uuid.UUID, roleId uuid.UUID) error {
	err := s.RolePermissionRepository.DeleteByRole(roleId)
	if err != nil {
		return err
	}
	chatUsers, err := s.ChatUserRepository.FindByChatAndRole(chatId, roleId)
	if err != nil {
		return err
	}
	defaultRole, err := s.ChatRoleRepository.FindDefaultRole(chatId)
	if err != nil {
		return err
	}
	for _, chatUser := range chatUsers {
		chatUser.RoleId = defaultRole.RoleId
		err = s.ChatUserRepository.UpdateChatUser(chatUser)
		if err != nil {
			return err
		}
	}
	err = s.RoleRepository.DeleteRole(roleId)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) doSetRole(req *dto.SetRoleRequest) error {
	for _, targetUserId := range req.TargetUserIds {
		chatUser, err := s.ChatUserRepository.FindByChatAndUser(req.ChatId, targetUserId)
		if err != nil {
			return err
		}
		chatUser.RoleId = req.RoleId
		err = s.ChatUserRepository.UpdateChatUser(chatUser)
		if err != nil {
			return err
		}
	}
	return nil
}
