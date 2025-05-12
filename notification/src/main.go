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
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	slog.Debug("Config: ", "cfg", cfg)

	database.Init(cfg)
	db := database.DB
	redisClient := client.NewRedisClient(cfg)
	smtpClient := client.NewSmtpClient(cfg)
	authClient := client.NewAuthClient(cfg)
	chatClient := client.NewChatClient(cfg)
	subscriptionRepository := repository.NewSubscriptionRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(
		notificationRepository,
		subscriptionRepository,
		authClient,
		chatClient,
		redisClient,
		smtpClient,
	)
	go redisClient.SubscribeToNotificationChannel(notificationService.SendNotification)
	go redisClient.SubscribeToSubscriptionChannel(notificationService.Subscribe)
	notificationController := controller.NewNotificationController(notificationService, authClient)
	httpServer := server.NewNotificationHttpServer(notificationController)
	app := app.New(httpServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
