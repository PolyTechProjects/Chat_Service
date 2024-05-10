package websocket

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"example.com/main/src/internal/websocket"
)

type App struct {
	log  *slog.Logger
	port int
}

func New(log *slog.Logger, port int) *App {
	return &App{
		log:  log,
		port: port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "websocketapp.Run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	websocket.SetupServer()
	log.Info("WebSocket server is running", slog.String("addr", l.Addr().String()))
	if err := http.Serve(l, nil); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
