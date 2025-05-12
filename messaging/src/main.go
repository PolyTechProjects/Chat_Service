package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/messaging/src/config"
	"example.com/messaging/src/database"
	"example.com/messaging/src/internal/app"
	"example.com/messaging/src/internal/client"
	"example.com/messaging/src/internal/controller"
	"example.com/messaging/src/internal/repository"
	"example.com/messaging/src/internal/server"
	"example.com/messaging/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	slog.Debug("Config: ", "cfg", cfg)

	database.Init(cfg)
	db := database.DB
	authClient := client.NewAuthClient(cfg)
	chatClient := client.NewChatClient(cfg)
	redisClient := client.NewRedisClient(cfg)
	messageRepository := repository.New(db)
	messageHistoryService := service.NewMessageHistoryService(messageRepository, chatClient)
	messageHistoryController := controller.NewMessageHistoryController(messageHistoryService, authClient)

	messageService := service.NewMessageService(messageRepository, chatClient, redisClient)
	go redisClient.SubscribeToMessageChannel(messageService.BroadcastMessageRedisChannel)
	webSocketController := controller.NewWebsocketController(messageService, authClient, chatClient)
	server := server.NewHttpServer(messageHistoryController, webSocketController)
	app := app.New(server, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer database.Close()
}
