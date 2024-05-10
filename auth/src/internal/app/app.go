package app

import (
	"log/slog"

	grpcapp "example.com/main/src/internal/app/grpc"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
) *App {
	log.Info("Starting gRPC server...")
	grpcApp := grpcapp.New(log, port)
	return &App{
		GRPCSrv: grpcApp,
	}
}
