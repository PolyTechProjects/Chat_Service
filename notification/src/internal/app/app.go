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
	httpPort   int
}

func New(httpServer *server.NotificationHttpServer, cfg *config.Config) *App {
	return &App{
		httpServer: httpServer,
		httpPort:   cfg.App.HttpInnerPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *App) Run() error {
	return a.RunHttpServer()
}

func (a *App) RunHttpServer() error {
	hl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.httpPort))
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	a.httpServer.StartServer(mux)
	handler := a.httpServer.ConfigureCors(mux)
	slog.Debug("Starting HTTP server")
	slog.Debug(hl.Addr().String())
	if err := http.Serve(hl, handler); err != nil {
		return err
	}
	return nil
}
