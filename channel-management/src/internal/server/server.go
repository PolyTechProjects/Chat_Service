package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	channelMgmt "example.com/channel-management/src/gen/go/channel_mgmt"
	"example.com/channel-management/src/internal/client"
	"example.com/channel-management/src/internal/controller"
	"example.com/channel-management/src/internal/dto"
	"example.com/channel-management/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	channelMgmtController *controller.ChannelManagementController
}

func NewHttpServer(channelMgmtController *controller.ChannelManagementController) *HttpServer {
	return &HttpServer{channelMgmtController: channelMgmtController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /channel", h.channelMgmtController.CreateChannelHandler)
	http.HandleFunc("DELETE /channel", h.channelMgmtController.DeleteChannelHandler)
	http.HandleFunc("PUT /channel", h.channelMgmtController.UpdateChannelHandler)
	http.HandleFunc("POST /join", h.channelMgmtController.JoinChannelHandler)
	http.HandleFunc("POST /leave", h.channelMgmtController.LeaveChannelHandler)
	http.HandleFunc("POST /invite", h.channelMgmtController.InviteUserHandler)
	http.HandleFunc("POST /kick", h.channelMgmtController.KickUserHandler)
	http.HandleFunc("POST /admin", h.channelMgmtController.MakeAdminHandler)
	http.HandleFunc("DELETE /admin", h.channelMgmtController.DeleteAdminHandler)
	http.HandleFunc("GET /channel", h.channelMgmtController.GetChannelHandler)
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	channelMgmt.UnimplementedChannelManagementServer
	service    *service.ChannelManagementService
	authClient *client.AuthGRPCClient
}

func New(service *service.ChannelManagementService, authClient *client.AuthGRPCClient) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer: gRPCServer,
		service:    service,
		authClient: authClient,
	}
	channelMgmt.RegisterChannelManagementServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) CreateChannel(ctx context.Context, req *channelMgmt.CreateChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Create channel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.CreatorId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	creatorId, err := uuid.Parse(req.CreatorId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	createReq := &dto.CreateChannelRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatorId:   creatorId,
	}
	channelID, err := s.service.CreateChannel(createReq)
	if err != nil {
		slog.Error("CreateChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Create channel controller successful", "channelID", channelID)

	channelId, err := uuid.Parse(channelID)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) DeleteChannel(ctx context.Context, req *channelMgmt.DeleteChannelRequest) (*channelMgmt.DeleteChannelResponse, error) {
	slog.Info("Delete channel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	delReq := &dto.DeleteChannelRequest{
		ChannelId: channelId,
		UserId:    userId,
	}
	err = s.service.DeleteChannel(delReq)
	if err != nil {
		slog.Error("DeleteChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete channel controller successful", "channelID", req.ChannelId)
	return &channelMgmt.DeleteChannelResponse{}, nil
}

func (s *GRPCServer) UpdateChannel(ctx context.Context, req *channelMgmt.UpdateChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Update channel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	updateReq := &dto.UpdateChannelRequest{
		ChannelId:   channelId,
		Name:        req.Name,
		UserId:      userId,
		Description: req.Description,
	}
	err = s.service.UpdateChannel(updateReq)
	if err != nil {
		slog.Error("UpdateChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update channel controller successful", "channelID", req.ChannelId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) JoinChannel(ctx context.Context, req *channelMgmt.JoinChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Join channel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	joinReq := &dto.JoinChannelRequest{
		ChannelId: channelId,
		UserId:    userId,
	}
	err = s.service.JoinChannel(joinReq)
	if err != nil {
		slog.Error("JoinChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join channel controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) LeaveChannel(ctx context.Context, req *channelMgmt.LeaveChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Leave channel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	leaveReq := &dto.LeaveChannelRequest{
		ChannelId: channelId,
		UserId:    userId,
	}
	err = s.service.LeaveChannel(leaveReq)
	if err != nil {
		slog.Error("LeaveChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Leave channel controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) InviteUser(ctx context.Context, req *channelMgmt.InviteUserRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Invite user controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	requestingUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	inviteReq := &dto.InviteUserRequest{
		ChannelId:        channelId,
		UserId:           userId,
		RequestingUserId: requestingUserId,
	}
	err = s.service.InviteUser(inviteReq)
	if err != nil {
		slog.Error("InviteUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Invite user controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *channelMgmt.KickUserRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Kick user controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	requestingUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	kickReq := &dto.KickUserRequest{
		ChannelId:        channelId,
		UserId:           userId,
		RequestingUserId: requestingUserId,
	}
	err = s.service.KickUser(kickReq)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *channelMgmt.MakeAdminRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("MakeAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	requestingUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.AdminRequest{
		ChannelId:        channelId,
		UserId:           userId,
		RequestingUserId: requestingUserId,
	}
	err = s.service.MakeAdmin(adminReq)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChatWithAdmins error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
		AdminsIds:       channel.Admins,
	}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *channelMgmt.DeleteAdminRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("DeleteAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	requestingUserId, err := uuid.Parse(req.RequestingUserId)
	if err != nil {
		slog.Error("Invalid requesting user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	delReq := &dto.AdminRequest{
		ChannelId:        channelId,
		UserId:           userId,
		RequestingUserId: requestingUserId,
	}
	err = s.service.DeleteAdmin(delReq)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChatWithAdmins error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
		AdminsIds:       channel.Admins,
	}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *channelMgmt.IsAdminRequest) (*channelMgmt.IsAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	adminReq := &dto.IsAdminRequest{
		ChannelId: channelId,
		UserId:    userId,
	}
	isAdmin, err := s.service.IsAdmin(adminReq)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChannel(ctx context.Context, req *channelMgmt.GetChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("GetChannel controller started")
	_, err := s.authClient.PerformAuthorize(ctx, nil, req.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(getReq)
	if err != nil {
		slog.Error("GetChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("GetChannel controller successful", "channelID", req.ChannelId)
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId.String(),
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
		AdminsIds:       channel.Admins,
	}, nil
}
