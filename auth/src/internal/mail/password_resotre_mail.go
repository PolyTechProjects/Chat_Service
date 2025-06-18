package mail

import (
	"fmt"

	"example.com/main/src/internal/dto"
)

type PasswordRestoreMail struct {
}

func NewPasswordRestoreMail() PasswordRestoreMail {
	return PasswordRestoreMail{}
}

func (p PasswordRestoreMail) Build(emailTo string, body string) *dto.SendEmailEvent {
	return &dto.SendEmailEvent{
		To:      emailTo,
		Subject: "Password restore",
		Body:    body,
	}
}

func (p PasswordRestoreMail) BuildBody(opts MailBodyOpts) string {
	return fmt.Sprintf("Для продолжения операции восстановления пароля перейдите по следующей ссылке: /%v", opts.VerificationToken)
}
