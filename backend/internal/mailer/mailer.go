package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	Host string
	Port int
	From string
}

func NewMailer(host string, port int, from string) *Mailer {
	return &Mailer{
		Host: host,
		Port: port,
		From: from,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", m.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	return smtp.SendMail(addr, nil, m.From, []string{to}, []byte(msg))
}
