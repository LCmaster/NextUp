package mailer

import (
	"context"
	"log/slog"
)

// Mailer defines the interface for sending emails.
type Mailer interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// MockMailer is a mock implementation of the Mailer interface that logs to the console.
// It is intended to be used in development and testing environments where real email
// sending is not required.
type MockMailer struct{}

// NewMockMailer creates a new MockMailer.
func NewMockMailer() *MockMailer {
	return &MockMailer{}
}

// SendEmail logs the email contents to the standard logger.
func (m *MockMailer) SendEmail(ctx context.Context, to, subject, body string) error {
	slog.Info("=== EMAIL SENT ===")
	slog.Info("To: " + to)
	slog.Info("Subject: " + subject)
	slog.Info("Body: " + body)
	slog.Info("==================")
	return nil
}
