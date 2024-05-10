package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"

	httpApp "example.com/chat-app/src/internal/server/http"
	"example.com/chat-app/src/internal/server/websocket"
)

type App struct {
	log  *slog.Logger
	port int
	h    *httpApp.HttpService
	ws   *websocket.WebsocketService
}

func New(log *slog.Logger, port int, h *httpApp.HttpService, ws *websocket.WebsocketService) *App {
	return &App{
		log:  log,
		port: port,
		h:    h,
		ws:   ws,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "webApp.Run"
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	a.h.SetupServer()
	a.ws.SetupServer()
	log.Info("Web server is running", slog.String("addr", l.Addr().String()))
	if err := http.Serve(l, nil); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}
