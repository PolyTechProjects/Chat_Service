package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"example.com/users/src/gen/go/users"
	"example.com/users/src/internal/controller"
	"example.com/users/src/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type HttpServer struct {
	UsersController *controller.UsersController
}

func NewHttpServer(UsersController *controller.UsersController) *HttpServer {
	return &HttpServer{UsersController: UsersController}
}

func (h *HttpServer) StartServer(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/v1/users/profiles/{profileLink}", h.UsersController.UpdateProfileHandler)
	mux.HandleFunc("GET /api/v1/users/profiles", h.UsersController.GetProfilesHandler)
	//mux.HandleFunc("GET /api/v1/users/profiles/{profileLink}", h.UsersController.GetProfileHandler)
}

func (h *HttpServer) ConfigureCors(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

type GRPCServer struct {
	gRPCServer *grpc.Server
	users.UnimplementedUsersServer
	usersService *service.UsersService
}

func NewGrpcServer(usersService *service.UsersService) *GRPCServer {
	gRPCServer := grpc.NewServer()
	g := &GRPCServer{
		gRPCServer:   gRPCServer,
		usersService: usersService,
	}
	users.RegisterUsersServer(gRPCServer, g)
	return g
}

func (s *GRPCServer) Start(l net.Listener) error {
	slog.Debug("Starting gRPC server")
	slog.Debug(l.Addr().String())
	return s.gRPCServer.Serve(l)
}

func (s *GRPCServer) GetUser(ctx context.Context, req *users.GetUserRequest) (*users.UserResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("GetUser failed: " + err.Error())
		return nil, err
	}
	user, err := s.usersService.GetUserById(userId)
	if err != nil {
		slog.Error("GetUser failed: " + err.Error())
		return nil, err
	}
	return &users.UserResponse{
		UserId:      user.UserId,
		Name:        user.Name,
		Firstname:   user.Firstname,
		Lastname:    user.Lastname,
		ProfilePic:  user.ProfilePic,
		ProfileLink: user.ProfileLink,
		Description: user.Description,
	}, nil
}

func (s *GRPCServer) GetUsers(ctx context.Context, req *users.GetUsersRequest) (*users.UsersResponse, error) {
	userIds := make([]uuid.UUID, len(req.UserIds))
	for i, user := range req.UserIds {
		userId, err := uuid.Parse(user)
		slog.Debug(userId.String())
		if err != nil {
			slog.Error("GetUsers failed: " + err.Error())
			return nil, err
		}
		userIds[i] = userId
	}
	usersByIds, err := s.usersService.GetUsersByIds(userIds)
	slog.Debug("UsersByIds", "count", len(usersByIds.Users))
	if err != nil {
		slog.Error("UsersServiceGetUsers failed: " + err.Error())
		return nil, err
	}
	response := make([]*users.UserResponse, len(usersByIds.Users))
	for i, user := range usersByIds.Users {
		slog.Debug("User", "user_id", user.UserId, "name", user.Name, "firstname", user.Firstname, "lastname", user.Lastname, "profile_pic", user.ProfilePic, "profile_link", user.ProfileLink, "description", user.Description)
		response[i] = &users.UserResponse{
			UserId:      user.UserId,
			Name:        user.Name,
			Firstname:   user.Firstname,
			Lastname:    user.Lastname,
			ProfilePic:  user.ProfilePic,
			ProfileLink: user.ProfileLink,
			Description: user.Description,
		}
	}
	return &users.UsersResponse{Users: response}, nil
}
