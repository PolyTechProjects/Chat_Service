package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"example.com/main/src/gen/go/auth"
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

func (h *HttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.authController.RegisterHandler)
	mux.HandleFunc("POST /api/v1/auth/login", h.authController.LoginHandler)
	mux.HandleFunc("POST /api/v1/auth/logout", h.authController.LogoutHandler)
	mux.HandleFunc("DELETE /api/v1/auth/{userId}", h.authController.DeleteAccountHandler)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.authController.RefreshHandler)
	mux.HandleFunc("GET /api/v1/auth/me", h.authController.MeHandler)
}

func (h *HttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	auth.UnimplementedAuthServer
	authService *service.AuthService
}

func New(authService *service.AuthService) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:  gRPCServer,
		authService: authService,
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
	authHeader := md.Get("Authorization")
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

func (s *GRPCServer) ExtractUserId(ctx context.Context, req *auth.ExtractUserIdRequest) (*auth.ExtractUserIdResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}
	authHeader := md.Get("Authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no auth header")
	}
	accessToken := strings.TrimPrefix(authHeader[0], "Bearer ")

	userId, err := s.authService.ExtractUserId(accessToken)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &auth.ExtractUserIdResponse{UserId: userId}, nil
}

func (s *GRPCServer) GetLogin(ctx context.Context, req *auth.GetLoginRequest) (*auth.GetLoginResponse, error) {
	login, err := s.authService.GetLogin(req.UserId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &auth.GetLoginResponse{Login: login}, nil
}
