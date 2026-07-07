package mailer

import (
	"context"
	"fmt"
	"net/smtp"
)

// SMTPMailer uses standard net/smtp to send emails.
type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPMailer creates a new SMTPMailer.
func NewSMTPMailer(host, port, username, password, from string) *SMTPMailer {
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// SendEmail connects to the SMTP server and sends an HTML email.
func (m *SMTPMailer) SendEmail(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	
	// Set up authentication information.
	var auth smtp.Auth
	if m.username != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	// Compose the MIME message
	header := make(map[string]string)
	header["From"] = m.from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
