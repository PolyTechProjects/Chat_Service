package app

import (
	"log/slog"

	grpcapp "example.com/main/src/internal/app/grpc"
	wsapp "example.com/main/src/internal/app/websocket"
)

type App struct {
	GRPCApp      *grpcapp.App
	WebSocketApp *wsapp.App
}

func New(log *slog.Logger, grpcPort int, wsPort int) *App {
	grpcApp := grpcapp.New(log, grpcPort)
	websocketApp := wsapp.New(log, wsPort)
	return &App{
		GRPCApp:      grpcApp,
		WebSocketApp: websocketApp,
	}
}
