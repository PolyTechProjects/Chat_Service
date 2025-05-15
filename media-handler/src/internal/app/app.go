package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/media/src/config"
	"example.com/media/src/internal/server"
)

type App struct {
	httpServer *server.HttpServer
	httpPort   int
}

func New(httpServer *server.HttpServer, cfg *config.Config) *App {
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
	go a.RunHttpServer()
	return nil
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
