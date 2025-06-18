package mail

type DefaultMailBuilder struct {
}

func NewDefaultMailBuilder() *DefaultMailBuilder {
	return &DefaultMailBuilder{}
}

func (b *DefaultMailBuilder) Build(emailTo string, subject string, body string) *Mail {
	return &Mail{
		To:      emailTo,
		Subject: subject,
		Body:    body,
	}
}

func (b *DefaultMailBuilder) BuildSubject(opts *MailSubjectOpts) string {
	return ""
}

func (b *DefaultMailBuilder) BuildBody(opts *MailBodyOpts) string {
	return ""
}
