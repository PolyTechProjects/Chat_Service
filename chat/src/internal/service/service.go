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
	UsersClient              *client.UsersGRPCClient
	RedisClient              *client.RedisClient
}

func NewChatService(
	chatRepository *repository.ChatRepository,
	directChatRepository *repository.DirectChatRepository,
	chatUserRepository *repository.ChatUserRepository,
	roleRepository *repository.RoleRepository,
	rolePermissionRepository *repository.RolePermissionRepository,
	usersClient *client.UsersGRPCClient,
	redisClient *client.RedisClient,
) *ChatService {
	return &ChatService{
		ChatRepository:           chatRepository,
		DirectChatRepository:     directChatRepository,
		ChatUserRepository:       chatUserRepository,
		RoleRepository:           roleRepository,
		RolePermissionRepository: rolePermissionRepository,
		UsersClient:              usersClient,
		RedisClient:              redisClient,
	}
}

func (s *ChatService) GetChatAndUserNames(chatId uuid.UUID, userId uuid.UUID) (string, string, error) {
	chat, err := s.ChatRepository.FindById(chatId)
	if err != nil {
		return "", "", err
	}
	user, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return "", "", err
	}
	return chat.Name, user.Nickname, nil
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
	err = s.VerifyUserAction(chatId, userId, models.CAN_DELETE_CHAT)
	if err != nil {
		return err
	}
	return s.doDeleteChat(chatId)
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
		Id:             uuid.New(),
		Name:           "default",
		Description:    "Default role",
		Color:          "#000000",
		ChatId:         chat.Id,
		IsDefault:      true,
		BasedOnDefault: true,
	}
	err = s.RoleRepository.AddRole(defaultRole)
	if err != nil {
		return nil, err
	}
	rolePermissions := make([]*models.RolePermission, len(req.DefaultPermissions))
	for i, permission := range req.DefaultPermissions {
		rolePermission := &models.RolePermission{
			RoleId:     defaultRole.Id,
			Permission: models.Permission(permission),
		}
		rolePermissions[i] = rolePermission
		err = s.RolePermissionRepository.AddRolePermission(rolePermission)
		if err != nil {
			return nil, err
		}
	}
	//err = s.RolePermissionRepository.AddRolePermissions(rolePermissions)

	defaultAdminRole := &models.Role{
		Id:             uuid.New(),
		Name:           "admin",
		Description:    "Admin role",
		Color:          "#ffffff",
		ChatId:         chat.Id,
		IsAdmin:        true,
		BasedOnDefault: false,
	}
	err = s.RoleRepository.AddRole(defaultAdminRole)
	if err != nil {
		return nil, err
	}
	defaultAdminPermissions := getAllPermissions()
	adminRolePermissions := make([]*models.RolePermission, len(defaultAdminPermissions))
	for i, permission := range defaultAdminPermissions {
		rolePermission := &models.RolePermission{
			RoleId:     defaultAdminRole.Id,
			Permission: permission,
		}
		adminRolePermissions[i] = rolePermission
		err = s.RolePermissionRepository.AddRolePermission(rolePermission)
		if err != nil {
			return nil, err
		}
	}
	//err = s.RolePermissionRepository.AddRolePermissions(adminRolePermissions)

	userIds := append(req.ParticipantsIds, req.CreatorId)
	participants := make([]*dto.Participants, len(userIds))
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
		err = s.ChatUserRepository.AddChatUser(chatUser)
		if err != nil {
			return nil, err
		}
		event := &dto.NewSubscriptionNotificationEvent{
			ChatId: chat.Id,
			UserId: participantId,
		}
		err = s.RedisClient.SendToSubscriptionChannel(event)
		if err != nil {
			return nil, err
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
	err = s.ChatUserRepository.AddChatUser(creator)
	//err = s.ChatUserRepository.AddChatUsers(users)
	if err != nil {
		return nil, err
	}
	event := &dto.NewSubscriptionNotificationEvent{
		ChatId: chat.Id,
		UserId: req.CreatorId,
	}
	err = s.RedisClient.SendToSubscriptionChannel(event)
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
	err = s.VerifyUserAction(req.ChatId, userId, models.CAN_EDIT_CHAT)
	if err != nil {
		return err
	}
	return s.doEditChat(req)
}

func (s *ChatService) JoinChat(joinLink string, userId uuid.UUID) (*dto.GetChatResponse, error) {
	chat, err := s.ChatRepository.FindByJoinLink(joinLink)
	if err != nil {
		return nil, err
	}
	defaultRole, err := s.RoleRepository.FindDefaultRole(chat.Id)
	if err != nil {
		return nil, err
	}
	chatUser := &models.ChatUser{
		ChatId:   chat.Id,
		UserId:   userId,
		RoleId:   defaultRole.Id,
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
	err = s.VerifyUserAction(req.ChatId, userId, models.CAN_ADD_USERS)
	if err != nil {
		return err
	}
	return s.doAddUsers(req)
}

func (s *ChatService) DeleteUsers(chatId uuid.UUID, userIds []uuid.UUID, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	for _, targetUserId := range userIds {
		err = s.VerifyUserActionOnSomebody(chatId, userId, targetUserId, models.CAN_DELETE_USERS)
		if err != nil {
			return err
		}
	}
	return s.doDeleteUsers(chatId, userIds)
}

func (s *ChatService) ChangeUserNickname(req *dto.ChangeUserNicknameRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	if req.UserId == userId {
		err = s.VerifyUserAction(req.ChatId, userId, models.CAN_CHANGE_OWN_NICKNAME)
		if err != nil {
			return err
		}
	} else {
		err = s.VerifyUserActionOnSomebody(req.ChatId, userId, req.UserId, models.CAN_CHANGE_OTHERS_NICKNAME)
		if err != nil {
			return err
		}
	}
	return s.doChangeUserNickname(req)
}

func (s *ChatService) CreateRole(req *dto.CreateRoleRequest, userId uuid.UUID) (*dto.RoleResponse, error) {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	if !userInChat {
		return nil, fmt.Errorf("user is not in chat")
	}
	err = s.VerifyUserAction(req.ChatId, userId, models.CAN_CREATE_ROLE)
	if err != nil {
		return nil, err
	}
	return s.doCreateRole(req)
}

func (s *ChatService) EditRole(req *dto.UpdateRoleRequest, userId uuid.UUID) (*dto.RoleResponse, error) {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return nil, err
	}
	if !userInChat {
		return nil, fmt.Errorf("user is not in chat")
	}
	err = s.VerifyUserAction(req.ChatId, userId, models.CAN_EDIT_ROLE)
	if err != nil {
		return nil, err
	}
	err = s.VerifyUserActionOnRole(req.ChatId, userId, req.RoleId, models.CAN_EDIT_ROLE)
	if err != nil {
		return nil, err
	}
	return s.doEditRole(req)
}

func (s *ChatService) DeleteRole(chatId uuid.UUID, userId uuid.UUID, roleId uuid.UUID) error {
	userInChat, err := s.isUserInChat(chatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	err = s.VerifyUserAction(chatId, userId, models.CAN_DELETE_ROLE)
	if err != nil {
		return err
	}
	err = s.VerifyUserActionOnRole(chatId, userId, roleId, models.CAN_DELETE_ROLE)
	if err != nil {
		return err
	}
	return s.doDeleteRole(chatId, roleId)
}

func (s *ChatService) SetRole(req *dto.SetRoleRequest, userId uuid.UUID) error {
	userInChat, err := s.isUserInChat(req.ChatId, userId)
	if err != nil {
		return err
	}
	if !userInChat {
		return fmt.Errorf("user is not in chat")
	}
	for _, targetUserId := range req.TargetUserIds {
		err = s.VerifyUserActionOnSomebody(req.ChatId, userId, targetUserId, models.CAN_SET_ROLES)
		if err != nil {
			return err
		}
	}
	return s.doSetRole(req)
}

func (s *ChatService) GetDirectChat(firstUserId uuid.UUID, secondUserId uuid.UUID) (*dto.DirectChatResponse, error) {
	firstUserId, secondUserId = s.sortUUIDs(firstUserId, secondUserId)
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

func (s *ChatService) CreateDirectChat(firstUserId uuid.UUID, secondUserId uuid.UUID) (*dto.DirectChatResponse, error) {
	firstUserId, secondUserId = s.sortUUIDs(firstUserId, secondUserId)
	directChat := &models.DirectChat{
		FirstUserId:  firstUserId,
		SecondUserId: secondUserId,
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

func (s *ChatService) LeaveFromChats(req *dto.LeaveFromChatsRequest, userId uuid.UUID) error {
	for i, request := range req.Request {
		if req.Request[i].IsDirect {
			firstUserId, secondUserId := s.sortUUIDs(userId, request.ChatId)
			directChat, err := s.DirectChatRepository.FindByUsers(firstUserId, secondUserId)
			if err != nil {
				return err
			}
			err = s.DirectChatRepository.DeleteDirectChat(directChat)
			if err != nil {
				return err
			}
		} else {
			chatUser, err := s.ChatUserRepository.FindByChatAndUser(request.ChatId, userId)
			if err != nil {
				return err
			}
			err = s.ChatUserRepository.DeleteChatUser(chatUser)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ChatService) GetChats(userId uuid.UUID) (*dto.ChatsResponse, error) {
	chatUsers, err := s.ChatUserRepository.FindByUser(userId)
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

	directChats, err := s.DirectChatRepository.FindByUser(userId)
	if err != nil {
		return nil, err
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

func (s *ChatService) VerifyUserAction(chatId uuid.UUID, userId uuid.UUID, action models.Permission) error {
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	roleId := chatUser.RoleId
	rolePermission, err := s.RolePermissionRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	for _, permission := range rolePermission {
		if permission.Permission == action {
			return nil
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) VerifyUserPersistance(chatId uuid.UUID, userId uuid.UUID) error {
	_, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) VerifyUserActionOnSomebody(chatId, userId, targetUserId uuid.UUID, action models.Permission) error {
	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	roleId := chatUser.RoleId
	role, err := s.RoleRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	rolePermission, err := s.RolePermissionRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	for _, permission := range rolePermission {
		if permission.Permission == action {
			targetChatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, targetUserId)
			if err != nil {
				return err
			}
			targetRole, err := s.RoleRepository.FindByRole(targetChatUser.RoleId)
			if err != nil {
				return err
			}
			if targetRole.BasedOnDefault || role.IsAdmin {
				return nil
			}
		}
	}
	return fmt.Errorf("permission denied")
}

func (s *ChatService) VerifyUserActionOnRole(chatId, userId, roleId uuid.UUID, action models.Permission) error {
	targetRole, err := s.RoleRepository.FindByRole(roleId)
	if err != nil {
		return err
	}
	if action == models.CAN_DELETE_ROLE && (targetRole.IsDefault || targetRole.IsAdmin) {
		return fmt.Errorf("cannot delete default/admin role")
	}

	chatUser, err := s.ChatUserRepository.FindByChatAndUser(chatId, userId)
	if err != nil {
		return err
	}
	role, err := s.RoleRepository.FindByRole(chatUser.RoleId)
	if err != nil {
		return err
	}
	if role.IsAdmin || targetRole.BasedOnDefault {
		return nil
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
	chatRoles, err := s.RoleRepository.FindByChat(chatId)
	if err != nil {
		return err
	}
	err = s.RoleRepository.DeleteByChat(chatId)
	if err != nil {
		return err
	}
	for _, chatRole := range chatRoles {
		err = s.RolePermissionRepository.DeleteByRole(chatRole.Id)
		if err != nil {
			return err
		}
		err = s.RoleRepository.DeleteRole(chatRole.Id)
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
	defaultRole, err := s.RoleRepository.FindDefaultRole(req.ChatId)
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
			RoleId:   defaultRole.Id,
			Nickname: users.Users[i].Name,
		}
		event := &dto.NewSubscriptionNotificationEvent{
			ChatId: req.ChatId,
			UserId: userId,
		}
		err = s.RedisClient.SendToSubscriptionChannel(event)
		if err != nil {
			return err
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
	baseRole, err := s.RoleRepository.FindById(req.BaseRoleId)
	if err != nil {
		return nil, err
	}
	newRole := &models.Role{
		Id:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		Color:          req.Color,
		ChatId:         req.ChatId,
		IsDefault:      false,
		IsAdmin:        false,
		BasedOnDefault: baseRole.BasedOnDefault,
	}
	err = s.RoleRepository.AddRole(newRole)
	if err != nil {
		return nil, err
	}
	newRolePermissions := make([]*models.RolePermission, len(req.Permissions))
	for i, permission := range req.Permissions {
		rolePermission := &models.RolePermission{
			Permission: models.Permission(permission),
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
			Permission: models.Permission(permission),
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
	defaultRole, err := s.RoleRepository.FindDefaultRole(chatId)
	if err != nil {
		return err
	}
	for _, chatUser := range chatUsers {
		chatUser.RoleId = defaultRole.Id
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
		role, err := s.RoleRepository.FindById(req.RoleId)
		if err != nil {
			return err
		}
		if role.IsAdmin {
			return fmt.Errorf("cannot set admin role")
		}
		chatUser.RoleId = req.RoleId
		err = s.ChatUserRepository.UpdateChatUser(chatUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ChatService) sortUUIDs(firstUserId uuid.UUID, secondUserId uuid.UUID) (uuid.UUID, uuid.UUID) {
	if firstUserId.String() > secondUserId.String() {
		firstUserId, secondUserId = secondUserId, firstUserId
	}
	return firstUserId, secondUserId
}
