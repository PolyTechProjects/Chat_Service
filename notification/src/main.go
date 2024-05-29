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
	"example.com/notification/src/internal/repository"
	"example.com/notification/src/internal/server"
	"example.com/notification/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	repository := repository.NewUserIdXDeviceTokenRepository(db)
	service := service.NewNotificationService(repository, cfg)
	client := client.NewUserMgmtClient(cfg)
	server := server.NewNotificationServer(service, client)
	app := app.New(server, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
