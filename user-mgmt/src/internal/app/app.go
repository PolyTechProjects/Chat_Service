package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/user-mgmt/src/config"
	"example.com/user-mgmt/src/internal/server"
)

type App struct {
	httpServer     *server.HttpServer
	UserMgmtServer *server.UserMgmtGRPCServer
	httpPort       int
	grpcPort       int
}

func New(httpServer *server.HttpServer, userMgmtServer *server.UserMgmtGRPCServer, cfg *config.Config) *App {
	return &App{
		httpServer:     httpServer,
		UserMgmtServer: userMgmtServer,
		httpPort:       cfg.App.HttpInnerPort,
		grpcPort:       cfg.App.GrpcInnerPort,
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
	gl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))
	if err != nil {
		return err
	}
	a.UserMgmtServer.Start(gl)
	return nil
}
