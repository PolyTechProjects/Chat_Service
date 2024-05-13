package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/main/src/config"
	"example.com/main/src/database"
	"example.com/main/src/internal/app"
	"example.com/main/src/internal/repository"
	"example.com/main/src/internal/server"
	"example.com/main/src/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	repository := repository.New(db)
	service := service.New(repository)
	server := server.New(service)
	app := app.New(log, cfg.GRPC.Port, server)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
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
