package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/database"
	"example.com/user-mgmt/src/internal/app"
	"example.com/user-mgmt/src/internal/client"
	"example.com/user-mgmt/src/internal/controller"
	"example.com/user-mgmt/src/internal/repository"
	"example.com/user-mgmt/src/internal/server"
	"example.com/user-mgmt/src/internal/service"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)
	database.Init(cfg)
	db := database.DB
	authClient := client.NewAuthClient(cfg)
	mediaHandlerClient := client.NewMediaHandlerClient(cfg)
	repository := repository.New(db)
	service := service.New(repository)
	controller := controller.New(service, authClient, mediaHandlerClient)
	httpServer := server.NewHttpServer(controller)
	grpcServer := server.NewGRPCServer(service, authClient)
	app := app.New(httpServer, grpcServer, cfg)
	go app.MustRun()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	defer log.Info("Program successfully finished!")
	defer db.Close()
}
