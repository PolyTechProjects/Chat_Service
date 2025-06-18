package mail

import "fmt"

type DirectNewMessageMailBuilder struct {
}

func NewDirectNewMessageMailBuilder() *DirectNewMessageMailBuilder {
	return &DirectNewMessageMailBuilder{}
}

func (b *DirectNewMessageMailBuilder) Build(emailTo string, subject string, body string) *Mail {
	return &Mail{
		To:      emailTo,
		Subject: subject,
		Body:    body,
	}
}

func (b *DirectNewMessageMailBuilder) BuildSubject(opts *MailSubjectOpts) string {
	return fmt.Sprintf("New direct message from %v", opts.SenderName)
}

func (b *DirectNewMessageMailBuilder) BuildBody(opts *MailBodyOpts) string {
	if opts.FilesCount > 0 {
		return fmt.Sprintf("You have received a new message from %v with %v files in: \n\"%v\"", opts.SenderName, opts.FilesCount, opts.Body)
	} else {
		return fmt.Sprintf("You have received a new message from %v: \n\"%v\"", opts.SenderName, opts.Body)
	}
}
