package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/chat-management/src/config"
	"example.com/chat-management/src/database"
	"example.com/chat-management/src/internal/app"
	"example.com/chat-management/src/internal/client"
	"example.com/chat-management/src/internal/controller"
	"example.com/chat-management/src/internal/repository"
	"example.com/chat-management/src/internal/server"
	"example.com/chat-management/src/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)

	log.Info("Initializing database connection")
	database.Init(cfg)
	defer database.Close()

	log.Info("Creating repository")
	repo := repository.NewChatRepository(database.DB)

	log.Info("Creating service")
	service := service.New(*repo)

	slog.Info("Creating auth client")
	authClient := client.NewAuthClient(cfg)

	slog.Info("Creating controller")
	controller := controller.NewChatManagementController(service, authClient)

	slog.Info("Creating gRPC server")
	grpcServer := server.New(service, authClient)

	slog.Info("Creating http server")
	httpServer := server.NewHttpServer(controller)

	log.Info("Starting application")
	application := app.New(httpServer, grpcServer, cfg)
	go application.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
