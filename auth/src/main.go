package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/main/src/config"
	"example.com/main/src/database"
	"example.com/main/src/internal/app"
	"example.com/main/src/internal/client"
	"example.com/main/src/internal/repository"
	"example.com/main/src/internal/server"
	"example.com/main/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	repository := repository.New(db)
	service := service.New(repository)
	client := client.New(cfg)
	server := server.New(service, client)
	app := app.New(server, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
