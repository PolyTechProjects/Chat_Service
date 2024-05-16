package app

import (
	"fmt"
	"log/slog"
	"net"

	"example.com/main/src/internal/server"
)

type GRPCApp struct {
	log        *slog.Logger
	port       int
	GRPCServer *server.GRPCServer
}

func New(log *slog.Logger, port int, gRPCServer *server.GRPCServer) *GRPCApp {
	return &GRPCApp{log: log, port: port, GRPCServer: gRPCServer}
}

func (a *GRPCApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *GRPCApp) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}
	if err = a.GRPCServer.Start(l); err != nil {
		return err
	}
	return nil
}
