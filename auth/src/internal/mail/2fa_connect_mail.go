package mail

import (
	"fmt"

	"example.com/main/src/internal/dto"
)

type TwoFactorAuthConnectMail struct {
}

func NewTwoFactorAuthConnectMail() TwoFactorAuthConnectMail {
	return TwoFactorAuthConnectMail{}
}

func (t TwoFactorAuthConnectMail) Build(emailTo string, body string) *dto.SendEmailEvent {
	return &dto.SendEmailEvent{
		To:      emailTo,
		Subject: "2FA connect",
		Body:    body,
	}
}

func (t TwoFactorAuthConnectMail) BuildBody(opts MailBodyOpts) string {
	return fmt.Sprintf("Код для подключения двухфакторной аутентификации: /%v", opts.TwoFactorAuthCode)
}
