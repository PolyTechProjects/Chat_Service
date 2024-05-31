package server

import (
	"context"
	"log/slog"
	"net"

	chatMgmt "example.com/chat-management/src/gen/go/chat-mgmt"
	auth "example.com/chat-management/src/gen/go/sso"
	"example.com/chat-management/src/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	gRPCServer *grpc.Server
	chatMgmt.UnimplementedChatManagementServer
	service    *service.ChatManagementService
	authClient auth.AuthClient
}

func New(service *service.ChatManagementService, authAddress string) (*GRPCServer, error) {
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
	chatMgmt.RegisterChatManagementServer(gRPCServer, g)
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

func (s *GRPCServer) CreateChat(ctx context.Context, req *chatMgmt.CreateChatRequest) (*chatMgmt.CreateChatResponse, error) {
	slog.Info("Create chat controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	chatID, err := s.service.CreateChat(ctx, req.Name, req.Description, req.CreatorId)
	if err != nil {
		slog.Error("CreateChat error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Create chat controller successful", "chatID", chatID)
	return &chatMgmt.CreateChatResponse{ChatId: chatID}, nil
}

func (s *GRPCServer) DeleteChat(ctx context.Context, req *chatMgmt.DeleteChatRequest) (*chatMgmt.DeleteChatResponse, error) {
	slog.Info("Delete chat controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.DeleteChat(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("DeleteChat error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Delete chat controller successful", "chatID", req.ChatId)
	return &chatMgmt.DeleteChatResponse{}, nil
}

func (s *GRPCServer) UpdateChat(ctx context.Context, req *chatMgmt.UpdateChatRequest) (*chatMgmt.UpdateChatResponse, error) {
	slog.Info("Update chat controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.UpdateChat(ctx, req.ChatId, req.Name, req.Description, req.UserId)
	if err != nil {
		slog.Error("UpdateChat error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Update chat controller successful", "chatID", req.ChatId)
	return &chatMgmt.UpdateChatResponse{}, nil
}

func (s *GRPCServer) JoinChat(ctx context.Context, req *chatMgmt.JoinChatRequest) (*chatMgmt.JoinChatResponse, error) {
	slog.Info("Join chat controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.JoinChat(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("JoinChat error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Join chat controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.JoinChatResponse{}, nil
}

func (s *GRPCServer) KickUser(ctx context.Context, req *chatMgmt.KickUserChatRequest) (*chatMgmt.KickUserChatResponse, error) {
	slog.Info("Kick user controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.KickUser(ctx, req.ChatId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("KickUser error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("Kick user controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.KickUserChatResponse{}, nil
}

func (s *GRPCServer) CanWrite(ctx context.Context, req *chatMgmt.CanWriteChatRequest) (*chatMgmt.CanWriteChatResponse, error) {
	slog.Info("CanWrite controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	canWrite, err := s.service.CanWrite(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("CanWrite error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("CanWrite controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.CanWriteChatResponse{CanWrite: canWrite}, nil
}

func (s *GRPCServer) MakeAdmin(ctx context.Context, req *chatMgmt.MakeChatAdminRequest) (*chatMgmt.MakeChatAdminResponse, error) {
	slog.Info("MakeAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.MakeAdmin(ctx, req.ChatId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("MakeAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("MakeAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.MakeChatAdminResponse{}, nil
}

func (s *GRPCServer) DeleteAdmin(ctx context.Context, req *chatMgmt.DeleteChatAdminRequest) (*chatMgmt.DeleteChatAdminResponse, error) {
	slog.Info("DeleteAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	err := s.service.DeleteAdmin(ctx, req.ChatId, req.UserId, req.RequestingUserId)
	if err != nil {
		slog.Error("DeleteAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("DeleteAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.DeleteChatAdminResponse{}, nil
}

func (s *GRPCServer) IsAdmin(ctx context.Context, req *chatMgmt.IsChatAdminRequest) (*chatMgmt.IsChatAdminResponse, error) {
	slog.Info("IsAdmin controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	isAdmin, err := s.service.IsAdmin(ctx, req.ChatId, req.UserId)
	if err != nil {
		slog.Error("IsAdmin error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("IsAdmin controller successful", "chatID", req.ChatId, "userID", req.UserId)
	return &chatMgmt.IsChatAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *GRPCServer) GetChatUsers(ctx context.Context, req *chatMgmt.GetChatUsersRequest) (*chatMgmt.GetChatUsersResponse, error) {
	slog.Info("GetChatUsers controller started")
	if err := s.authorize(ctx); err != nil {
		slog.Error("Authorization error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	userIDs, err := s.service.GetChatUsers(ctx, req.ChatId)
	if err != nil {
		slog.Error("GetChatUsers error", "error", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	slog.Info("GetChatUsers controller successful", "chatID", req.ChatId)
	return &chatMgmt.GetChatUsersResponse{UserIds: userIDs}, nil
}
