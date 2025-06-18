package mail

type Mail struct {
	To      string
	Subject string
	Body    string
}

type MailBuilder interface {
	Build(emailTo string, subject string, body string) *Mail
	BuildSubject(opts *MailSubjectOpts) string
	BuildBody(opts *MailBodyOpts) string
}

type MailSubjectOpts struct {
	ChatName   string
	SenderName string
}

type MailBodyOpts struct {
	SenderName string
	FilesCount int
	Body       string
}
