package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/main/src/config"
	"example.com/main/src/database"
	"example.com/main/src/internal/app"
	"example.com/main/src/models"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	application := app.New(log, cfg.GRPC.Port, cfg.WebSocket.Port)
	go application.GRPCApp.MustRun()
	go application.WebSocketApp.MustRun()
	fmt.Println(cfg)
	database.Init()
	db := database.DB
	db.AutoMigrate(&models.User{})
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("DB :%b", db.HasTable(&models.User{}))
	defer log.Info("Program successfully finished!")
	defer db.Close()
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
