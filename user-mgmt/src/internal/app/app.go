package app

import (
	"fmt"
	"net"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/internal/server"
)

type GRPCApp struct {
	UserMgmtServer *server.UserMgmtGRPCServer
	Port           int
}

func New(userMgmtServer *server.UserMgmtGRPCServer, cfg *config.Config) *GRPCApp {
	return &GRPCApp{
		UserMgmtServer: userMgmtServer,
		Port:           cfg.App.InnerPort,
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
	if err = a.UserMgmtServer.Start(l); err != nil {
		return err
	}
	return nil
}
