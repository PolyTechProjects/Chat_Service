package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/user_mgmt/src/config"
	"example.com/user_mgmt/src/database"
	"example.com/user_mgmt/src/internal/app"
	"example.com/user_mgmt/src/internal/client"
	"example.com/user_mgmt/src/internal/controller"
	"example.com/user_mgmt/src/internal/repository"
	"example.com/user_mgmt/src/internal/server"
	"example.com/user_mgmt/src/internal/service"
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
	app := app.New(httpServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
