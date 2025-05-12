package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/users/src/config"
	"example.com/users/src/database"
	"example.com/users/src/internal/app"
	"example.com/users/src/internal/client"
	"example.com/users/src/internal/controller"
	"example.com/users/src/internal/repository"
	"example.com/users/src/internal/server"
	"example.com/users/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	authClient := client.NewAuthClient(cfg)
	redisClient := client.NewRedisClient(cfg)
	repository := repository.New(db)
	service := service.New(repository)
	go redisClient.SubscribeToCreateAccountChannel(service)
	go redisClient.SubscribeToDeleteAccountChannel(service)
	controller := controller.New(service, authClient)
	httpServer := server.NewHttpServer(controller)
	grpcServer := server.NewGrpcServer(service)
	app := app.New(httpServer, grpcServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
