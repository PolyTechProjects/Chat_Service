package mail

import (
	"fmt"

	"example.com/main/src/internal/dto"
)

type EmailVerificationMail struct {
}

func NewEmailVerificationMail() EmailVerificationMail {
	return EmailVerificationMail{}
}

func (e EmailVerificationMail) Build(emailTo string, body string) *dto.SendEmailEvent {
	return &dto.SendEmailEvent{
		To:      emailTo,
		Subject: "Email verification",
		Body:    body,
	}
}

func (e EmailVerificationMail) BuildBody(opts MailBodyOpts) string {
	return fmt.Sprintf("Для подтверждения почты перейдите по следующей ссылке: /%v", opts.VerificationToken)
}
