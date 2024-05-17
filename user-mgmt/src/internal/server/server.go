package server

import (
	"context"
	"log/slog"
	"net"

	"example.com/user-mgmt/src/gen/go/sso"
	user_mgmt "example.com/user-mgmt/src/gen/go/user-mgmt"
	"example.com/user-mgmt/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	user_mgmt.UnimplementedUserMgmtServer
	userMgmtService *service.UserMgmtService
	client          sso.AuthClient
}

func New(userMgmtService *service.UserMgmtService, ssoUrl string) *GRPCServer {
	conn, err := grpc.NewClient(ssoUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	client := sso.NewAuthClient(conn)
	gRPCServcer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:      gRPCServcer,
		userMgmtService: userMgmtService,
		client:          client,
	}
	user_mgmt.RegisterUserMgmtServer(g.gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) InfoUpdate(ctx context.Context, req *user_mgmt.InfoUpdateRequest) (*user_mgmt.UserResponse, error) {
	authResp, err := s.client.Authorize(ctx, &sso.AuthorizeRequest{Token: req.Token})
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
