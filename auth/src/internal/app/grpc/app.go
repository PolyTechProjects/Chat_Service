package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	"example.com/main/src/internal/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	auth.Register(gRPCServer)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (app *App) Run() {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", app.port))
	if err != nil {
		app.log.Error("failed to listen: %v", err)
	}
	app.gRPCServer.Serve(l)
}

func (app *App) Stop() {
	app.log.Info("Stopping gRPC server...")
	app.gRPCServer.GracefulStop()
}
