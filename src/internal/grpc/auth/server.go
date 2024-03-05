package auth

import (
	"context"

	"google.golang.org/grpc"

	// Сгенерированный код

	ssov1 "example.com/main/src/gen/go/sso"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer // Хитрая штука, о ней ниже
}

// Тот самый интерфейс, котрый мы передавали в grpcApp
type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
}

func Register(gRPCServer *grpc.Server) {
	ssov1.RegisterAuthServer(gRPCServer, &serverAPI{})
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	panic("implement me")
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	panic("implement me")
}
