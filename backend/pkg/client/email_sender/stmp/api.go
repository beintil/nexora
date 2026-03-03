package email_stmp

import (
	"context"
	"net/smtp"
	"telephony/pkg/client/email_sender"
)

type SMTPSender struct {
	host string
	port string
	from string
	auth smtp.Auth
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host: host,
		port: port,
		from: from,
		auth: smtp.PlainAuth("", username, password, host),
	}
}

func (s *SMTPSender) Send(ctx context.Context, msg email_sender.Message) error {
	body := buildMessage(s.from, msg)
	return smtp.SendMail(
		s.host+":"+s.port,
		s.auth,
		s.from,
		[]string{msg.To},
		[]byte(body),
	)
}

func buildMessage(from string, msg email_sender.Message) string {
	headers := ""
	headers += "MIME-version: 1.0;\r\n"
	headers += "Content-Type: text/html; charset=\"UTF-8\";\r\n"
	headers += "From: " + from + "\r\n"
	headers += "To: " + msg.To + "\r\n"
	headers += "Subject: " + msg.Subject + "\r\n\r\n"

	return headers + msg.HTML
}
