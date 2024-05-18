package server

import (
	"context"
	"log/slog"
	"net"

	user_mgmt "example.com/user-mgmt/src/gen/go/user-mgmt"
	"example.com/user-mgmt/src/internal/client"
	"example.com/user-mgmt/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserMgmtGRPCServer struct {
	gRPCServer *grpc.Server
	user_mgmt.UnimplementedUserMgmtServer
	userMgmtService *service.UserMgmtService
	authClient      *client.AuthGRPCClient
}

func New(userMgmtService *service.UserMgmtService, authClient *client.AuthGRPCClient) *UserMgmtGRPCServer {
	gRPCServcer := grpc.NewServer()
	g := &UserMgmtGRPCServer{
		gRPCServer:      gRPCServcer,
		userMgmtService: userMgmtService,
		authClient:      authClient,
	}
	user_mgmt.RegisterUserMgmtServer(g.gRPCServer, g)
	return g
}

func (s *UserMgmtGRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *UserMgmtGRPCServer) AddUser(ctx context.Context, req *user_mgmt.AddUserRequest) (*user_mgmt.UserResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.userMgmtService.CreateUser(userId, req.Name)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp := &user_mgmt.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Description: "",
		Avatar:      "",
	}
	return resp, nil
}

func (s *UserMgmtGRPCServer) InfoUpdate(ctx context.Context, req *user_mgmt.InfoUpdateRequest) (*user_mgmt.UserResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, req.Token)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if !authResp.Authorized {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.userMgmtService.UpdateUser(userId, req.Name, req.Description)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp := &user_mgmt.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Description: user.Description,
		Avatar:      user.Avatar,
	}
	return resp, nil
}
