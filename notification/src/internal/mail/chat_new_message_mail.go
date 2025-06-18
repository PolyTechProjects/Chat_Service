package mail

import "fmt"

type ChatNewMessageMailBuilder struct {
}

func NewChatNewMessageMailBuiler() *ChatNewMessageMailBuilder {
	return &ChatNewMessageMailBuilder{}
}

func (b *ChatNewMessageMailBuilder) Build(emailTo string, subject string, body string) *Mail {
	return &Mail{
		To:      emailTo,
		Subject: subject,
		Body:    body,
	}
}

func (b *ChatNewMessageMailBuilder) BuildSubject(opts *MailSubjectOpts) string {
	return fmt.Sprintf("%v: New message from %v", opts.ChatName, opts.SenderName)
}

func (b *ChatNewMessageMailBuilder) BuildBody(opts *MailBodyOpts) string {
	if opts.FilesCount > 0 {
		return fmt.Sprintf("You have received a new message from %v with %v files in: \n\"%v\"", opts.SenderName, opts.FilesCount, opts.Body)
	} else {
		return fmt.Sprintf("You have received a new message from %v: \n\"%v\"", opts.SenderName, opts.Body)
	}
}
