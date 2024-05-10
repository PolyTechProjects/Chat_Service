package websocket

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/main/src/internal/dto"
	"github.com/gorilla/websocket"
)

var clients = make(map[*Client]bool)
var broadcastChannel = make(chan *dto.MessageRequest)

type Client struct {
	wsConnection *websocket.Conn
	Username     string
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error has occured while trying to connect to websocket server.")
	}
	slog.Info("WebSocket Connection opened")
	client := &Client{
		wsConnection: wsConnection,
		Username:     "user",
	}
	clients[client] = true

	readMessages(client)

	delete(clients, client)
	wsConnection.Close()
}

func readMessages(client *Client) {
	for {
		_, payload, err := client.wsConnection.ReadMessage()
		if err != nil {
			slog.Error("Error has occured while reading message", err)
			return
		}

		message := dto.MessageRequest{}

		err = json.Unmarshal(payload, &message)
		if err != nil {
			slog.Error("Error has occured while unmarshalling message", err)
			return
		}

		broadcastChannel <- &message
	}
}

func broadcast() {
	for {
		message := <-broadcastChannel
		slog.Info("New message:\t" + message.Body)
		for client := range clients {
			err := client.wsConnection.WriteJSON(message)
			if err != nil {
				slog.Error("Error has occured while sending message", err)
				client.wsConnection.Close()
				delete(clients, client)
			}
		}
	}
}

func setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Simple Server")
	})
	http.HandleFunc("/websocket", serveWs)
}

func SetupServer() {
	go broadcast()
	setupRoutes()
}
