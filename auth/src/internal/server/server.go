package server

import (
	"context"
	"log/slog"
	"net"

	"example.com/main/src/gen/go/sso"
	"example.com/main/src/internal/client"
	"example.com/main/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	sso.UnimplementedAuthServer
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
	sso.RegisterAuthServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	token, err := s.authService.Register(req.GetLogin(), req.GetUsername(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.userMgmtClient.PerformAddUser(ctx, req.GetLogin(), req.GetUsername())
	return &sso.RegisterResponse{Token: token}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	token, err := s.authService.Login(req.GetLogin(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &sso.LoginResponse{Token: token}, nil
}

func (s *GRPCServer) Authorize(ctx context.Context, req *sso.AuthorizeRequest) (*sso.AuthorizeResponse, error) {
	err := s.authService.Authorize(req.GetToken())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &sso.AuthorizeResponse{Authorized: true}, nil
}
