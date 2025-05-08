package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/media/src/config"
	"example.com/media/src/database"
	"example.com/media/src/internal/app"
	"example.com/media/src/internal/client"
	"example.com/media/src/internal/controller"
	"example.com/media/src/internal/repository"
	"example.com/media/src/internal/server"
	"example.com/media/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	authClient := client.NewAuthClient(cfg)
	redisClient := client.NewRedisClient(cfg)
	seaweedFSClient := client.NewSeaweedFSCLient(cfg)
	repository := repository.New(db)
	service := service.New(repository, redisClient, seaweedFSClient)
	controller := controller.New(service, authClient)
	httpServer := server.NewHttpServer(controller)
	app := app.New(httpServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	database.Close()
}
