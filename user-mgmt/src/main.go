package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/database"
	"example.com/user-mgmt/src/internal/app"
	"example.com/user-mgmt/src/internal/repository"
	"example.com/user-mgmt/src/internal/server"
	"example.com/user-mgmt/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	repository := repository.New(db)
	service := service.New(repository)
	ssoConnString := fmt.Sprintf("%s:%s", os.Getenv("AUTH_HOST"), os.Getenv("AUTH_PORT"))
	server := server.New(service, ssoConnString)
	app := app.New(cfg.GRPCConfig.Port, server)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
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
