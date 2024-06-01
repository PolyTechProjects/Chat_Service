package server

import (
	"context"
	"log/slog"
	"net"

	channelMgmt "example.com/channel-management/src/gen/go/channel-mgmt"
	"example.com/channel-management/src/internal/client"
	"example.com/channel-management/src/internal/service"
	"google.golang.org/grpc"
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

func (s *GRPCServer) CreateChannel(ctx context.Context, req *channelMgmt.CreateChannelRequest) (*channelMgmt.CreateChannelResponse, error) {
	slog.Info("Create channel controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	channelID, err := s.service.CreateChannel(ctx, req.Name, req.Description, req.CreatorId)
	if err != nil {
		slog.Error("CreateChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Create channel controller successful", "channelID", channelID)
	return &channelMgmt.CreateChannelResponse{ChannelId: channelID}, nil
}

func (s *GRPCServer) DeleteChannel(ctx context.Context, req *channelMgmt.DeleteChannelRequest) (*channelMgmt.DeleteChannelResponse, error) {
	slog.Info("Delete channel controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.DeleteChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("DeleteChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Delete channel controller successful", "channelID", req.ChannelId)
	return &channelMgmt.DeleteChannelResponse{}, nil
}

func (s *GRPCServer) UpdateChannel(ctx context.Context, req *channelMgmt.UpdateChannelRequest) (*channelMgmt.UpdateChannelResponse, error) {
	slog.Info("Update channel controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.UpdateChannel(ctx, req.ChannelId, req.Name, req.Description, req.UserId)
	if err != nil {
		slog.Error("UpdateChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Update channel controller successful", "channelID", req.ChannelId)
	return &channelMgmt.UpdateChannelResponse{}, nil
}

func (s *GRPCServer) JoinChannel(ctx context.Context, req *channelMgmt.JoinChannelRequest) (*channelMgmt.JoinChannelResponse, error) {
	slog.Info("Join channel controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.JoinChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("JoinChannel error", "error", err.Error())
		return nil, err
	}
	slog.Info("Join channel controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.JoinChannelResponse{}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *channelMgmt.KickUserChannelRequest) (*channelMgmt.KickUserChannelResponse, error) {
	slog.Info("Kick user controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.KickUser(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, err
	}
	slog.Info("Kick user controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.KickUserChannelResponse{}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *channelMgmt.CanWriteChannelRequest) (*channelMgmt.CanWriteChannelResponse, error) {
	slog.Info("CanWrite controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	canWrite, err := s.service.CanWrite(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, err
	}
	slog.Info("CanWrite controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.CanWriteChannelResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *channelMgmt.MakeChannelAdminRequest) (*channelMgmt.MakeChannelAdminResponse, error) {
	slog.Info("MakeAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.MakeAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("MakeAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.MakeChannelAdminResponse{}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *channelMgmt.DeleteChannelAdminRequest) (*channelMgmt.DeleteChannelAdminResponse, error) {
	slog.Info("DeleteAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	err := s.service.DeleteAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("DeleteAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.DeleteChannelAdminResponse{}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *channelMgmt.IsChannelAdminRequest) (*channelMgmt.IsChannelAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	isAdmin, err := s.service.IsAdmin(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, err
	}
	slog.Info("IsAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &channelMgmt.IsChannelAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChanUsers(ctx context.Context, req *channelMgmt.GetChanUsersRequest) (*channelMgmt.GetChanUsersResponse, error) {
	slog.Info("GetChanUsers controller started")
	if err := s.authClient.PerformAuthorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, err
	}
	userIDs, err := s.service.GetChanUsers(ctx, req.ChannelId)
	if err != nil {
		slog.Error("GetChanUsers error", "error", err.Error())
		return nil, err
	}
	slog.Info("GetChanUsers controller successful", "channelID", req.ChannelId)
	return &channelMgmt.GetChanUsersResponse{UserIds: userIDs}, nil
}
