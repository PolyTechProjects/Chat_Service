package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	userMgmt "example.com/user-mgmt/src/gen/go/user_mgmt"
	"example.com/user-mgmt/src/internal/client"
	"example.com/user-mgmt/src/internal/controller"
	"example.com/user-mgmt/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HttpServer struct {
	userMgmtController *controller.UserMgmtController
}

func NewHttpServer(userMgmtController *controller.UserMgmtController) *HttpServer {
	return &HttpServer{userMgmtController: userMgmtController}
}

func (h *HttpServer) StartServer() {
	http.HandleFunc("POST /upload", h.userMgmtController.UpdateAvatarHandler)
	http.HandleFunc("PUT /info", h.userMgmtController.InfoUpdateHandler)
	http.HandleFunc("GET /user", h.userMgmtController.GetUserHandler)
	http.HandleFunc("DELETE /user", h.userMgmtController.DeleteUserHandler)
}

type UserMgmtGRPCServer struct {
	gRPCServer *grpc.Server
	userMgmt.UnimplementedUserMgmtServer
	userMgmtService *service.UserMgmtService
	authClient      *client.AuthGRPCClient
}

func NewGRPCServer(userMgmtService *service.UserMgmtService, authClient *client.AuthGRPCClient) *UserMgmtGRPCServer {
	gRPCServcer := grpc.NewServer()
	g := &UserMgmtGRPCServer{
		gRPCServer:      gRPCServcer,
		userMgmtService: userMgmtService,
		authClient:      authClient,
	}
	userMgmt.RegisterUserMgmtServer(g.gRPCServer, g)
	return g
}

func (s *UserMgmtGRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *UserMgmtGRPCServer) AddUser(ctx context.Context, req *userMgmt.AddUserRequest) (*userMgmt.UserResponse, error) {
	slog.Info(fmt.Sprintf("Add User %v", req.UserId))
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
	resp := &userMgmt.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Description: "",
		Avatar:      "",
	}
	return resp, nil
}

func (s *UserMgmtGRPCServer) InfoUpdate(ctx context.Context, req *userMgmt.InfoUpdateRequest) (*userMgmt.UserResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userId, err := uuid.Parse(req.GetUserId())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.userMgmtService.UpdateUser(userId, req.Name, req.Description)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp := &userMgmt.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Description: user.Description,
		Avatar:      user.Avatar,
	}
	return resp, nil
}

func (s *UserMgmtGRPCServer) GetUser(ctx context.Context, req *userMgmt.GetUserRequest) (*userMgmt.UserResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userId, err := uuid.Parse(req.GetUserId())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.userMgmtService.GetUser(userId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp := &userMgmt.UserResponse{
		UserId:      user.Id.String(),
		Name:        user.Name,
		Description: user.Description,
		Avatar:      user.Avatar,
	}
	return resp, nil
}

func (s *UserMgmtGRPCServer) DeleteAccount(ctx context.Context, req *userMgmt.DeleteAccountRequest) (*userMgmt.DummyResponse, error) {
	authResp, err := s.authClient.PerformAuthorize(ctx, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if authResp.UserId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "Unauthorized")
	}

	userId, err := uuid.Parse(req.GetUserId())
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.userMgmtService.DeleteUser(userId)
	if err != nil {
		slog.Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &userMgmt.DummyResponse{}, nil
}
