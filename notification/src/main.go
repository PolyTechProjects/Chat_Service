package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/notification/src/config"
	"example.com/notification/src/database"
	"example.com/notification/src/internal/app"
	"example.com/notification/src/internal/client"
	"example.com/notification/src/internal/controller"
	"example.com/notification/src/internal/repository"
	"example.com/notification/src/internal/server"
	"example.com/notification/src/internal/service"
	"example.com/notification/src/redis"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	redis.Init(cfg)
	redis := redis.RedisClient
	repository := repository.NewUserIdXDeviceTokenRepository(db, redis)
	service := service.NewNotificationService(repository, cfg)
	authClient := client.NewAuthClient(cfg)
	userMgmtClient := client.NewUserMgmtClient(cfg)
	controller := controller.NewNotificationController(service, authClient, userMgmtClient)
	httpServer := server.NewNotificationHttpServer(controller)
	gRPCServer := server.NewNotificationGRPCServer(service, authClient, userMgmtClient)
	app := app.New(httpServer, gRPCServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
