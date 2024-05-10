package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"example.com/chat-app/src/internal/dto"
	"example.com/chat-app/src/internal/models"
	"example.com/chat-app/src/internal/repository"
)

type HttpService struct {
	Repository *repository.Repository
}

func (h *HttpService) getHistory(chatRoomId uint64) (*dto.HistoryResponse, error) {
	messages, err := h.Repository.GetMessageByChatRoomId(chatRoomId)
	if err != nil {
		return nil, err
	}
	var messageResps []dto.MessageResponse

	for _, message := range messages {
		messageResps = append(messageResps, *models.MapMessageToResponse(&message))
	}
	return &dto.HistoryResponse{
		Messages: messageResps,
	}, nil
}

func (h *HttpService) getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	chatRoomId, err := strconv.ParseUint(strings.Split(r.URL.Path, "/")[1], 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat room id", http.StatusBadRequest)
		return
	}
	history, err := h.getHistory(chatRoomId)
	if err != nil {
		http.Error(w, "No such chat room", http.StatusNotFound)
		return
	}
	response, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}
	w.Write(response)
}

func (h *HttpService) setupRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Simple Server")
	})
	http.HandleFunc("/{chatRoomId}/history/", h.getHistoryHandler)
}

func (h *HttpService) SetupServer() {
	h.setupRoutes()
}
