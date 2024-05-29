package app

import (
	"fmt"
	"log/slog"
	"net"

	"example.com/chat-management/src/internal/server"
)

type App struct {
	log  *slog.Logger
	port int
	srv  *server.GRPCServer
}

func New(log *slog.Logger, port int, srv *server.GRPCServer) *App {
	return &App{log: log, port: port, srv: srv}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}
	a.log.Info("Starting gRPC server", slog.String("address", l.Addr().String()))
	return a.srv.Start(l)
}
