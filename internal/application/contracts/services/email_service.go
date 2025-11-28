package services

import "context"

type EmailService interface {
	SendEmail(ctx context.Context, input SendEmailInput) error
	SendVerificationEmail(ctx context.Context, email, username, token string) error
	SendPasswordResetEmail(ctx context.Context, email, username, token string) error
	SendWelcomeEmail(ctx context.Context, email, username string) error
}

type SendEmailInput struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}
