package dto

type MessageRequest struct {
	SenderId   string `json:"senderId"`
	ChatRoomId string `json:"chatRoomId"`
	Body       string `json:"body"`
	CreatedAt  uint64 `json:"createdAt"`
	WithMedia  bool   `json:"withMedia"`
}

type MessageResponse struct {
	SenderId   string   `json:"senderId"`
	ChatRoomId string   `json:"chatRoomId"`
	Body       string   `json:"body"`
	CreatedAt  uint64   `json:"createdAt"`
	Metadata   Metadata `json:"metadata"`
}

type Metadata struct {
	FilePath string
}

func MapRequestToResponse(req MessageRequest) *MessageResponse {
	resp := &MessageResponse{}
	resp.Body = req.Body
	resp.ChatRoomId = req.ChatRoomId
	resp.CreatedAt = req.CreatedAt
	resp.SenderId = req.SenderId
	return resp
}

type HistoryResponse struct {
	Messages []MessageResponse `json:"messages"`
}
