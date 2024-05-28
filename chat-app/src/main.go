package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/chat-app/src/config"
	"example.com/chat-app/src/database"
	"example.com/chat-app/src/internal/app"
	"example.com/chat-app/src/internal/client"
	"example.com/chat-app/src/internal/controller"
	"example.com/chat-app/src/internal/repository"
	"example.com/chat-app/src/internal/server"
	"example.com/chat-app/src/internal/service"
	"example.com/chat-app/src/redis"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	redis.Init(cfg)
	redisClient := redis.RedisClient
	database.Init(cfg)
	db := database.DB
	authClient := client.NewAuthClient(cfg)
	messageRepository := repository.New(db, redisClient)
	messageHistoryService := service.NewMessageHistoryService(messageRepository)
	messageService := service.NewMessageService(messageRepository)
	messageHistoryController := controller.NewMessageHistoryController(messageHistoryService)
	webSocketController := controller.NewWebsocketController(messageService, authClient)
	server := server.NewHttpServer(messageHistoryController, webSocketController)
	app := app.New(server, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer database.Close()
	defer redis.Close()
}
