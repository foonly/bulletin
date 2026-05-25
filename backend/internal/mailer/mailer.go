package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type Mailer struct {
	Host     string
	Port     int
	From     string
	Username string
	Password string
	UseTLS   bool // For implicit TLS (port 465)
}

func NewMailer(host string, port int, from, username, password string, useTLS bool) *Mailer {
	return &Mailer{
		Host:     host,
		Port:     port,
		From:     from,
		Username: username,
		Password: password,
		UseTLS:   useTLS,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", m.From, to, subject, body))

	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	if m.UseTLS {
		// Implicit TLS (usually port 465)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         m.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, m.Host)
		if err != nil {
			return err
		}
		defer c.Quit()

		if m.Username != "" {
			auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
			if err = c.Auth(auth); err != nil {
				return err
			}
		}

		if err = c.Mail(m.From); err != nil {
			return err
		}
		if err = c.Rcpt(to); err != nil {
			return err
		}

		w, err := c.Data()
		if err != nil {
			return err
		}

		_, err = w.Write(msg)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}

		return nil
	}

	// Standard STARTTLS (usually port 587) or No Encryption
	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	return smtp.SendMail(addr, auth, m.From, []string{to}, msg)
}
