package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/notification/src/config"
	"example.com/notification/src/internal/server"
)

type App struct {
	httpServer *server.NotificationHttpServer
	gRPCServer *server.NotificationGRPCServer
	httpPort   int
	gRPCPort   int
}

func New(httpServer *server.NotificationHttpServer, gRPCServer *server.NotificationGRPCServer, cfg *config.Config) *App {
	return &App{
		httpServer: httpServer,
		httpPort:   cfg.App.HttpInnerPort,
		gRPCServer: gRPCServer,
		gRPCPort:   cfg.App.GRPCInnerPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *App) Run() error {
	go a.RunHttpServer()
	a.RunGRPCServer()
	return nil
}

func (a *App) RunHttpServer() error {
	hl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.httpPort))
	if err != nil {
		return err
	}
	slog.Debug("Starting HTTP server")
	slog.Debug(hl.Addr().String())
	a.httpServer.StartServer()
	if err := http.Serve(hl, nil); err != nil {
		return err
	}
	return nil
}

func (a *App) RunGRPCServer() error {
	gl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.gRPCPort))
	if err != nil {
		return err
	}
	a.gRPCServer.Start(gl)
	return nil
}
