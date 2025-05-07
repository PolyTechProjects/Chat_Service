package dto

type RegisterRequest struct {
	Login    string `json:"login"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserId   string `json:"user_id"`
	Login    string `json:"login"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AccountCreatedEvent struct {
	UserId   string `json:"user_id"`
	Login    string `json:"login"`
	Username string `json:"username"`
}
