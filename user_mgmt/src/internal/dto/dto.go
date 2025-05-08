package dto

type UpdateUserRequest struct {
	Name        string `json:"name"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	ProfilePic  string `json:"profile_pic"`
	ProfileLink string `json:"profile_link"`
	Description string `json:"description"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserResponse struct {
	UserId      string `json:"user_id"`
	Name        string `json:"name"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	ProfilePic  string `json:"profile_pic"`
	ProfileLink string `json:"profile_link"`
	Description string `json:"description"`
}

type DeleteUserRequest struct {
	UserId string `json:"user_id"`
}

type AccountCreatedEvent struct {
	UserId    string `json:"user_id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type AccountDeletedEvent struct {
	UserId string `json:"user_id"`
}
