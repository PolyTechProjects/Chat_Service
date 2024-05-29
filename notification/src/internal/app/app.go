package app

import (
	"fmt"
	"net"

	"example.com/notification/src/config"
	"example.com/notification/src/internal/server"
)

type App struct {
	NotificationServer *server.NotificationServer
	gRPCPort           int
}

func New(notificationServer *server.NotificationServer, cfg *config.Config) *App {
	return &App{
		NotificationServer: notificationServer,
		gRPCPort:           cfg.App.GRPCInnerPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.gRPCPort))
	if err != nil {
		return err
	}
	if err = a.NotificationServer.Start(l); err != nil {
		return err
	}
	return nil
}
