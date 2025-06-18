package mail

import "example.com/main/src/internal/dto"

type Mail interface {
	Build(emailTo string, body string) *dto.SendEmailEvent
	BuildBody(opts MailBodyOpts) string
}

type MailBodyOpts struct {
	VerificationToken string
	TwoFactorAuthCode string
}
