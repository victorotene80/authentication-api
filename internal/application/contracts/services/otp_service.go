package services

import (
	"context"
)


type OTPService interface {
	// Generate creates and stores an OTP for the given email
	Generate(ctx context.Context, email string, purpose string) (string, error)
	
	// Verify checks if the provided OTP is valid for the email and purpose
	Verify(ctx context.Context, email string, otpCode string, purpose string) (bool, error)
	
	// SendEmail sends OTP via email
	SendEmail(ctx context.Context, email string, otpCode string, purpose string) error
	
	// IsRateLimited checks if user has exceeded OTP request limits
	IsRateLimited(ctx context.Context, email string) (bool, error)
}

// OTPPurpose defines different OTP use cases
const (
	OTPPurposeLogin            = "login"
	OTPPurposeEmailVerification = "email_verification"
	OTPPurposePasswordReset    = "password_reset"
)

// OTPConfig holds OTP configuration
