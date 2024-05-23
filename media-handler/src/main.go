package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/media-handler/src/config"
	"example.com/media-handler/src/database"
	"example.com/media-handler/src/internal/app"
	"example.com/media-handler/src/internal/client"
	"example.com/media-handler/src/internal/controller"
	"example.com/media-handler/src/internal/repository"
	"example.com/media-handler/src/internal/server"
	"example.com/media-handler/src/internal/service"
	"example.com/media-handler/src/redis"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	redis.Init()
	redisClient := redis.RedisClient
	database.Init(cfg)
	db := database.DB
	authClient := client.New(cfg)
	repository := repository.New(db, redisClient)
	service := service.New(repository, cfg)
	controller := controller.New(service, authClient)
	httpServer := server.NewHttpServer(controller)
	grpcServer := server.NewGRPCServer(service, authClient)
	app := app.New(httpServer, grpcServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	database.Close()
	redis.Close()
}
