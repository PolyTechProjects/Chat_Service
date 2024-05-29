package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/channel-management/src/config"
	"example.com/channel-management/src/database"
	"example.com/channel-management/src/internal/app"
	"example.com/channel-management/src/internal/repository"
	"example.com/channel-management/src/internal/server"
	"example.com/channel-management/src/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)

	slog.Info("Initializing database connection")
	database.Init(cfg)
	defer database.Close()

	slog.Info("Creating repository")
	repo := repository.NewChannelRepository(database.DB)

	slog.Info("Creating service")
	service := service.New(*repo)

	slog.Info("Creating gRPC server")
	grpcServer, err := server.New(service, os.Getenv("AUTH_ADDRESS"))
	if err != nil {
		slog.Error("Failed to create gRPC server", "error", err)
		return
	}

	slog.Info("Starting application")
	application := app.New(cfg.Grpc.Port, grpcServer)
	go application.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
