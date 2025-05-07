package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"example.com/main/src/gen/go/auth"
	"example.com/main/src/internal/client"
	"example.com/main/src/internal/controller"
	"example.com/main/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	authController *controller.AuthController
}

func NewHttpServer(authController *controller.AuthController) *HttpServer {
	return &HttpServer{
		authController: authController,
	}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /register", h.authController.RegisterHandler)
	http.HandleFunc("POST /login", h.authController.LoginHandler)
	http.HandleFunc("POST /logout", h.authController.LogoutHandler)
}

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
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no auth header")
	}
	accessToken := strings.TrimPrefix(authHeader[0], "Bearer ")

	err := s.authService.Authorize(accessToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return nil, nil
}

func (s *GRPCServer) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	accessToken, refreshToken, err := s.authService.RefreshTokens(req.GetRefreshToken())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &auth.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

/*
	Access Token = 10 mins
	Refresh Token = 10 Days
	Client sends access token and receives access to resource
	If access token is expired, then server returns 401, so client sends refresh token to refresh access token
	When access token is refreshed, server also refreshes a refresh token and saves new version in database
	After that, server returns new access token and old refresh token
	So, when old refresh token expires, server will take refresh token from db, verify that it is not expired, refresh both tokens and return
*/
