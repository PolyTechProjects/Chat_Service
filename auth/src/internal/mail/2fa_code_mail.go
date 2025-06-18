package mail

import (
	"fmt"

	"example.com/main/src/internal/dto"
)

type TwoFactorAuthCodeMail struct {
}

func NewTwoFactorAuthCodeMail() TwoFactorAuthCodeMail {
	return TwoFactorAuthCodeMail{}
}

func (t TwoFactorAuthCodeMail) Build(emailTo string, body string) *dto.SendEmailEvent {
	return &dto.SendEmailEvent{
		To:      emailTo,
		Subject: "2FA code",
		Body:    body,
	}
}

func (t TwoFactorAuthCodeMail) BuildBody(opts MailBodyOpts) string {
	return fmt.Sprintf("Код для входа в аккаунт: /%v", opts.TwoFactorAuthCode)
}
