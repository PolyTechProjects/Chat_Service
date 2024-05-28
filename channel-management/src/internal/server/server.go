package server

import (
	"context"
	"log/slog"
	"net"

	sso "example.com/channel-management/src/gen/go/channel-mgmt"
	auth "example.com/channel-management/src/gen/go/sso"
	"example.com/channel-management/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	sso.UnimplementedChannelManagementServer
	service    *service.ChannelManagementService
	authClient auth.AuthClient
}

func New(service *service.ChannelManagementService, authAddress string) (*GRPCServer, error) {
	conn, err := grpc.Dial(authAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authClient := auth.NewAuthClient(conn)
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer: gRPCServer,
		service:    service,
		authClient: authClient,
	}
	sso.RegisterChannelManagementServer(gRPCServer, g)
	return g, nil
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Error("Missing metadata")
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	token := md["authorization"]
	if len(token) == 0 {
		slog.Error("Missing token")
		return status.Error(codes.Unauthenticated, "missing token")
	}

	req := &auth.AuthorizeRequest{Token: token[0]}
	res, err := s.authClient.Authorize(ctx, req)
	if err != nil {
		slog.Error("Authorization failed", "error", err.Error())
		return status.Error(codes.PermissionDenied, "unauthorized")
	}
	if !res.Authorized {
		slog.Error("Unauthorized access attempt")
		return status.Error(codes.PermissionDenied, "unauthorized")
	}
	return nil
}

func (s *GRPCServer) CreateChannel(ctx context.Context, req *sso.CreateChannelRequest) (*sso.CreateChannelResponse, error) {
	slog.Info("Create channel controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	channelID, err := s.service.CreateChannel(ctx, req.Name, req.Description, req.CreatorId)
	if err != nil {
		slog.Error("CreateChannel error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Create channel controller successful", "channelID", channelID)
	return &sso.CreateChannelResponse{ChannelId: channelID}, nil
}

func (s *GRPCServer) DeleteChannel(ctx context.Context, req *sso.DeleteChannelRequest) (*sso.DeleteChannelResponse, error) {
	slog.Info("Delete channel controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.DeleteChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("DeleteChannel error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Delete channel controller successful", "channelID", req.ChannelId)
	return &sso.DeleteChannelResponse{}, nil
}

func (s *GRPCServer) UpdateChannel(ctx context.Context, req *sso.UpdateChannelRequest) (*sso.UpdateChannelResponse, error) {
	slog.Info("Update channel controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.UpdateChannel(ctx, req.ChannelId, req.Name, req.Description, req.UserId)
	if err != nil {
		slog.Error("UpdateChannel error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Update channel controller successful", "channelID", req.ChannelId)
	return &sso.UpdateChannelResponse{}, nil
}

func (s *GRPCServer) JoinChannel(ctx context.Context, req *sso.JoinChannelRequest) (*sso.JoinChannelResponse, error) {
	slog.Info("Join channel controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.JoinChannel(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("JoinChannel error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Join channel controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.JoinChannelResponse{}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *sso.KickUserRequest) (*sso.KickUserResponse, error) {
	slog.Info("Kick user controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.KickUser(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Kick user controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.KickUserResponse{}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *sso.CanWriteRequest) (*sso.CanWriteResponse, error) {
	slog.Info("CanWrite controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	canWrite, err := s.service.CanWrite(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("CanWrite controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.CanWriteResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *sso.MakeAdminRequest) (*sso.MakeAdminResponse, error) {
	slog.Info("MakeAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.MakeAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("MakeAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.MakeAdminResponse{}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *sso.DeleteAdminRequest) (*sso.DeleteAdminResponse, error) {
	slog.Info("DeleteAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.DeleteAdmin(ctx, req.ChannelId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("DeleteAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.DeleteAdminResponse{}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *sso.IsAdminRequest) (*sso.IsAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	isAdmin, err := s.service.IsAdmin(ctx, req.ChannelId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("IsAdmin controller successful", "channelID", req.ChannelId, "userID", req.UserId)
	return &sso.IsAdminResponse{IsAdmin: isAdmin}, nil
}
