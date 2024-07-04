package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/channel-management/src/config"
	"example.com/channel-management/src/database"
	"example.com/channel-management/src/internal/app"
	"example.com/channel-management/src/internal/client"
	"example.com/channel-management/src/internal/controller"
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

	slog.Info("Creating auth client")
	authClient := client.NewAuthClient(cfg)

	slog.Info("Creating controller")
	controller := controller.NewChannelManagementController(service, authClient)

	slog.Info("Creating gRPC server")
	grpcServer := server.New(service, authClient)

	slog.Info("Creating HTTP server")
	httpServer := server.NewHttpServer(controller)

	slog.Info("Starting application")
	application := app.New(httpServer, grpcServer, cfg)
	go application.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
