package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/main/src/config"
	"example.com/main/src/internal/server"
)

type App struct {
	gRPCServer *server.GRPCServer
	gRPCPort   int
	httpServer *server.HttpServer
	httpPort   int
}

func New(gRPCServer *server.GRPCServer, httpServer *server.HttpServer, cfg *config.Config) *App {
	return &App{
		gRPCServer: gRPCServer,
		gRPCPort:   cfg.App.InnerGrpcPort,
		httpServer: httpServer,
		httpPort:   cfg.App.InnerHttpPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err.Error())
	}
}

func (a *App) Run() error {
	go a.RunGRPCServer()
	a.RunHttpServer()
	return nil
}

func (a *App) RunGRPCServer() error {
	gl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.gRPCPort))
	if err != nil {
		return err
	}
	slog.Debug("Starting gRPC server")
	slog.Debug(gl.Addr().String())
	if err = a.gRPCServer.Start(gl); err != nil {
		return err
	}
	return nil
}

func (a *App) RunHttpServer() error {
	hl, err := net.Listen("tcp", fmt.Sprintf(":%d", a.httpPort))
	if err != nil {
		slog.Error("Error has occured while listening: " + err.Error())
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
