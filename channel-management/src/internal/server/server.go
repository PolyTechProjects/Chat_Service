package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	channelMgmt "example.com/channel-management/src/gen/go/channel_mgmt"
	"example.com/channel-management/src/internal/client"
	"example.com/channel-management/src/internal/dto"
	"example.com/channel-management/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	channelMgmt.UnimplementedChannelManagementServer
	service    *service.ChannelManagementService
	authClient *client.AuthGRPCClient
}

func New(service *service.ChannelManagementService, authClient *client.AuthGRPCClient) (*GRPCServer, error) {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer: gRPCServer,
		service:    service,
		authClient: authClient,
	}
	channelMgmt.RegisterChannelManagementServer(gRPCServer, g)
	return g, nil
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) CreateChannel(ctx context.Context, req *channelMgmt.CreateChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Create channel controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.CreatorId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.CreatorId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	channelID, err := s.service.CreateChannel(ctx, req.Name, req.Description, req.CreatorId)
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
	channel, err := s.service.GetChannel(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) DeleteChannel(ctx context.Context, req *channelMgmt.DeleteChannelRequest) (*channelMgmt.DeleteChannelResponse, error) {
	slog.Info("Delete channel controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.DeleteChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("DeleteChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete channel controller successful", "channelID", req.ChannelId)
	return &channelMgmt.DeleteChannelResponse{}, nil
}

func (s *GRPCServer) UpdateChannel(ctx context.Context, req *channelMgmt.UpdateChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Update channel controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.UpdateChannel(ctx, req.ChannelId, req.Name, req.Description, req.UserId)
	if err != nil {
		slog.Error("UpdateChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update channel controller successful", "channelID", req.ChannelId)

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) JoinChannel(ctx context.Context, req *channelMgmt.JoinChannelRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Join channel controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.JoinChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("JoinChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join channel controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *channelMgmt.KickUserRequest) (*channelMgmt.ChannelResponse, error) {
	slog.Info("Kick user controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.KickUser(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannel(ctx, getReq)
	if err != nil {
		slog.Error("GetChat error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
	}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *channelMgmt.CanWriteRequest) (*channelMgmt.CanWriteResponse, error) {
	slog.Info("CanWrite controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	canWrite, err := s.service.CanWrite(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, err
	}
	slog.Info("CanWrite controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.CanWriteResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *channelMgmt.MakeAdminRequest) (*channelMgmt.ChannelWithAdminsResponse, error) {
	slog.Info("MakeAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.MakeAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannelWithAdmins(ctx, getReq)
	if err != nil {
		slog.Error("GetChatWithAdmins error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelWithAdminsResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
		AdminsIds:       channel.Admins,
	}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *channelMgmt.DeleteAdminRequest) (*channelMgmt.ChannelWithAdminsResponse, error) {
	slog.Info("DeleteAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	err = s.service.DeleteAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)

	channelId, err := uuid.Parse(req.ChannelId)
	if err != nil {
		slog.Error("Invalid channel Id", "error", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	getReq := &dto.GetChannelRequest{
		ChannelId: channelId,
	}
	channel, err := s.service.GetChannelWithAdmins(ctx, getReq)
	if err != nil {
		slog.Error("GetChatWithAdmins error", "error", err.Error())
		return nil, err
	}
	return &channelMgmt.ChannelWithAdminsResponse{
		ChannelId:       channel.Channel.Id.String(),
		CreatorId:       channel.Channel.CreatorId,
		Name:            channel.Channel.Name,
		Description:     channel.Channel.Description,
		ParticipantsIds: channel.Users,
		AdminsIds:       channel.Admins,
	}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *channelMgmt.IsAdminRequest) (*channelMgmt.IsAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	authResp, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}
	if authResp.UserId != req.UserId {
		err = fmt.Errorf("authorization error: %v and %v are not same", authResp.UserId, req.UserId)
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	isAdmin, err := s.service.IsAdmin(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChanUsers(ctx context.Context, req *channelMgmt.GetUsersRequest) (*channelMgmt.GetUsersResponse, error) {
	slog.Info("GetChanUsers controller started")
	_, err := s.authClient.PerformAuthorize(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Authorization error: %v", err.Error()))
		return nil, err
	}

	userIDs, err := s.service.GetChanUsers(ctx, req.ChannelId)
	if err != nil {
		slog.Error("GetChanUsers error", "error", err.Error())
		return nil, err
	}
	slog.Info("GetChanUsers controller successful", "channelID", req.ChannelId)
	return &channelMgmt.GetUsersResponse{UserIds: userIDs}, nil
}
