package auth

import (
	"context"

	"example.com/main/src/gen/go/sso"
	"google.golang.org/grpc"
)

type serverApi struct {
	sso.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	sso.RegisterAuthServer(gRPC, &serverApi{})
}

func (s *serverApi) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	panic("implement me")
}

func (s *serverApi) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	panic("implement me")
}
