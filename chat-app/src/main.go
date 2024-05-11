package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/chat-app/src/config"
	"example.com/chat-app/src/database"
	"example.com/chat-app/src/internal/app"
	"example.com/chat-app/src/internal/repository"
	"example.com/chat-app/src/internal/server/http"
	"example.com/chat-app/src/internal/server/websocket"
	"example.com/chat-app/src/redis"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	fmt.Println(cfg)
	database.Init(cfg)
	redis.InitRedis()
	repository := &repository.Repository{
		DB: database.DB,
	}
	httpService := &http.HttpService{
		Repository: repository,
	}
	websocketService := websocket.New(repository)
	webApp := app.New(log, cfg.Web.Port, httpService, websocketService)
	go webApp.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer database.Close()
	defer redis.Close()
	defer log.Info("Program successfully finished!")
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
