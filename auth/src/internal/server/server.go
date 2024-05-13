package server

import (
	"context"
	"log/slog"
	"net"

	"example.com/main/src/gen/go/sso"
	"example.com/main/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	sso.UnimplementedAuthServer
	AuthService *service.AuthService
}

func New(authService *service.AuthService) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{gRPCServer: gRPCServer, AuthService: authService}
	sso.RegisterAuthServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	token, err := s.AuthService.Register(req.GetLogin(), req.GetUsername(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &sso.RegisterResponse{Token: token}, nil
}

func (s *GRPCServer) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	token, err := s.AuthService.Login(req.GetLogin(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &sso.LoginResponse{Token: token}, nil
}
