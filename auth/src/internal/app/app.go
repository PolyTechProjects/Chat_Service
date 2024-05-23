package app

import (
	"fmt"
	"net"

	"example.com/main/src/config"
	"example.com/main/src/internal/server"
)

type GRPCApp struct {
	AuthServer *server.GRPCServer
	Port       int
}

func New(authServer *server.GRPCServer, cfg *config.Config) *GRPCApp {
	return &GRPCApp{
		AuthServer: authServer,
		Port:       cfg.App.InnerPort,
	}
}

func (a *GRPCApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *GRPCApp) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Port))
	if err != nil {
		return err
	}
	if err = a.AuthServer.Start(l); err != nil {
		return err
	}
	return nil
}
