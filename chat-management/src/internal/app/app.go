package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/chat-management/src/config"
	"example.com/chat-management/src/internal/server"
)

type App struct {
	httpServer *server.HttpServer
	httpPort   int
	gRPCServer *server.GRPCServer
	gRPCPort   int
}

func New(httpServer *server.HttpServer, gRPCServer *server.GRPCServer, cfg *config.Config) *App {
	return &App{httpServer: httpServer, gRPCServer: gRPCServer, httpPort: cfg.App.HttpInnerPort, gRPCPort: cfg.App.GrpcInnerPort}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
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
	slog.Info("Starting HTTP server")
	slog.Info(hl.Addr().String())
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
