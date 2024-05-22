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
	token, userId, err := s.authService.Register(req.GetLogin(), req.GetUsername(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = s.userMgmtClient.PerformAddUser(ctx, userId.String(), req.GetUsername())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
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

func (s *GRPCServer) ExtractUserId(ctx context.Context, req *sso.ExtractUserIdRequest) (*sso.ExtractUserIdResponse, error) {
	userId, err := s.authService.ExtractUserId(req.GetToken())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &sso.ExtractUserIdResponse{UserId: userId}, nil
}
