package app

import (
	"log/slog"
	"time"

	grpcapp "example.com/main/src/internal/app/grpc"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	log.Info("Starting gRPC server...")
	grpcApp := grpcapp.New(log, port)
	return &App{
		GRPCSrv: grpcApp,
	}
}
