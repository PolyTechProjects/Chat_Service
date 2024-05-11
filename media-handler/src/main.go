package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/media-handler/src/config"
	"example.com/media-handler/src/database"
	"example.com/media-handler/src/internal/app"
	"example.com/media-handler/src/internal/controller"
	"example.com/media-handler/src/internal/repository"
	"example.com/media-handler/src/internal/server"
	"example.com/media-handler/src/internal/service"
	"example.com/media-handler/src/redis"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	redis.Init()
	redisClient := redis.RedisClient
	database.Init(cfg)
	db := database.DB
	repository := repository.New(db, redisClient)
	service := service.New(repository, cfg.SeaweedFS.MasterIp, cfg.SeaweedFS.MasterPort)
	controller := controller.New(service)
	server := server.New(controller)
	app := app.New(log, cfg.Web.Port, server)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	database.Close()
	redis.Close()
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
