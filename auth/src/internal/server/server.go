package server

import (
	"context"
	"log/slog"
	"net"

	"example.com/main/src/gen/go/auth"
	"example.com/main/src/internal/client"
	"example.com/main/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	auth.UnimplementedAuthServer
	authService    *service.AuthService
	userMgmtClient *client.UserMgmtGRPCClient
}

func New(authService *service.AuthService, userMgmtClient *client.UserMgmtGRPCClient) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:     gRPCServer,
		authService:    authService,
		userMgmtClient: userMgmtClient,
	}
	auth.RegisterAuthServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	accessToken, refreshToken, userId, err := s.authService.Register(req.GetLogin(), req.GetUsername(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = s.userMgmtClient.PerformAddUser(ctx, userId.String(), req.GetUsername())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &auth.RegisterResponse{AccessToken: accessToken, RefreshToken: refreshToken, UserId: userId.String()}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	accessToken, refreshToken, err := s.authService.Login(req.GetLogin(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auth.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *GRPCServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	accessToken, userId, err := s.authService.Authorize(req.GetAccessToken(), req.GetRefreshToken())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &auth.AuthorizeResponse{AccessToken: accessToken, UserId: userId.String()}, nil
}
