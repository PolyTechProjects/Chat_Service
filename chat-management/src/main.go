package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/chat-management/src/config"
	"example.com/chat-management/src/database"
	"example.com/chat-management/src/internal/app"
	"example.com/chat-management/src/internal/repository"
	"example.com/chat-management/src/internal/server"
	"example.com/chat-management/src/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Initializing database connection")
	database.Init(cfg)
	defer database.Close()

	log.Info("Creating repository")
	repo := repository.NewChatRepository(database.DB)

	log.Info("Creating service")
	service := service.New(*repo)

	log.Info("Creating gRPC server")
	grpcServer, err := server.New(service, os.Getenv("AUTH_ADDRESS"))
	if err != nil {
		log.Error("Failed to create gRPC server", "error", err)
		return
	}

	log.Info("Starting application")
	application := app.New(log, cfg.Grpc.Port, grpcServer)
	go application.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
