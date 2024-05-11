package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/media-handler/src/internal/server"
)

type App struct {
	log        *slog.Logger
	port       int
	httpServer *server.HttpServer
}

func New(log *slog.Logger, port int, httpServer *server.HttpServer) *App {
	return &App{port: port, log: log, httpServer: httpServer}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}
	a.httpServer.StartServer()
	log.Info("MediaHandlerServer is running", slog.String("address", l.Addr().String()))
	if err := http.Serve(l, nil); err != nil {
		return err
	}
	return nil
}
