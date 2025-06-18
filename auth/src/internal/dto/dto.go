package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Login     string `json:"login"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Password  string `json:"password"`
}

type RegisterResponse struct {
	UserId    string `json:"user_id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type MeResponse struct {
	UserId uuid.UUID `json:"user_id"`
}

type VerifyEmailRequest struct {
	VerificationToken string `json:"verification_token"`
}

type ResetPasswordRequest struct {
	VerificationToken string `json:"verification_token"`
	NewPassword       string `json:"new_password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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

type SendEmailEvent struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
