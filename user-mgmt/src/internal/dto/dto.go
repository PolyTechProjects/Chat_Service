package dto

type UpdateInfoRequest struct {
	UserId      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateInfoResponse struct {
	UserId string `json:"user_id"`
}

type GetUserResponse struct {
	UserId      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

type DeleteUserRequest struct {
	UserId string `json:"user_id"`
}
