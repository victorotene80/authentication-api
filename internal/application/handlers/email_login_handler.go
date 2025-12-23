package handlers

import (
	"context"
	"errors"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/contracts/services"
	"authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	domainServices "authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type LoginEmailHandler struct {
	userRepo       repositories.UserRepository
	sessionRepo    repositories.SessionRepository
	outbox         persistence.OutboxRepository
	uow            persistence.UnitOfWork
	passwordHasher *domainServices.PasswordHashingService
	tokenService   services.TokenService
	otpService     services.OTPService
	logger         logging.Logger
}

func NewLoginEmailHandler(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	outbox persistence.OutboxRepository,
	uow persistence.UnitOfWork,
	passwordHasher *domainServices.PasswordHashingService,
	tokenService services.TokenService,
	otpService services.OTPService,
	logger logging.Logger,
) messaging.CommandHandler[commands.LoginEmailUserCommand, dtos.LoginEmailUserResult] {
	return &LoginEmailHandler{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		outbox:         outbox,
		uow:            uow,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
		otpService:     otpService,
		logger:         logger.With(zap.String("handler", "login_email")),
	}
}

func (h *LoginEmailHandler) Handle(
	ctx context.Context,
	cmd commands.LoginEmailUserCommand,
) (dtos.LoginEmailUserResult, error) {
	defer func(){
		if r := recover(); r != nil{
			h.logger.Error(ctx, "Panic during email user login",
				zap.Any("panic", r),
				zap.String("email", cmd.Email),
			)
			_ = h.publishFailedLoginEvent(ctx, cmd, fmt.Errorf("panic: %v", r))
			panic(r)
		}
	}()

	emailVO, err := valueobjects.NewEmail(cmd.Email)
	if err != nil {
		_ = h.publishFailedLoginEvent(ctx, cmd, err)
		return dtos.LoginEmailUserResult{}, err
	}

	user, err := h.userRepo.FindByEmail(ctx, emailVO)
	if err != nil {
		if err == domain.ErrUserNotFound {
			_ = h.publishFailedLoginEvent(ctx, cmd, err)
			return dtos.LoginEmailUserResult{}, domain.ErrInvalidCredentials
		}

		return dtos.LoginEmailUserResult{}, fmt.Errorf("error fetching user by email: %w", err)
	}

	if !user.User.IsActive{
		_ = h.publishFailedLoginEvent(ctx, cmd, domain.ErrUserInactive)
		return dtos.LoginEmailUserResult{}, domain.ErrUserInactive
	}

	if user.User.IsOAuthUser(){
		return h.handleOAuthUserLogin(ctx, user, cmd)
	}

	if !h.passwordHasher.Verify(cmd.Password, user.User.Password){
		_ = h.publishFailedLoginEvent(ctx, cmd, domain.ErrInvalidCredentials)
		return dtos.LoginEmailUserResult{}, domain.ErrInvalidCredentials
	}

	if !user.User.IsVerified{
		return dtos.LoginEmailUserResult{}, domain.ErrEmailNotVerified()
	}

	return h.generateTokensAndLogin(ctx, user, cmd)
}

func (h *LoginEmailHandler) handleOAuthUserLogin(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.LoginEmailUserCommand,
) (dtos.LoginEmailUserResult, error){

	isLimited, err := h.otpService.IsRateLimited(ctx, cmd.Email)
	if err != nil {
		return dtos.LoginEmailUserResult{}, fmt.Errorf("failed to check OTP rate limit: %w", err)
	}

	if isLimited {
		return dtos.LoginEmailUserResult{}, domain.ErrOTPRateLimited
	}

	otpCode, err := h.otpService.Generate(ctx, cmd.Email, services.OTPPurposeLogin)
	if err != nil{
		return dtos.LoginEmailUserResult{}, fmt.Errorf("failed to generate login OTP: %w", err)
	}

	if err := h.otpService.SendEmail(ctx, cmd.Email, otpCode, services.OTPPurposeLogin); err != nil {
		h.logger.Error(ctx, "Failed to send OTP email",
			zap.Error(err),
			zap.String("email", cmd.Email),
		)
		return dtos.LoginEmailUserResult{}, fmt.Errorf("failed to send OTP: %w", err)
	}

	return dtos.LoginEmailUserResult{
		Email: cmd.Email,
		RequiresOTP: true,
		OTPSent: true,
		Message: fmt.Sprintf("This account uses %s for sign-in. We've sent a verification code to your email.", user.User.Provider())
	}, nil

}


func (h *LoginEmailHandler) generateTokensAndLogin(
	ctx contect.Context,
	user *aggregates.UserAggregate,
	cmd commands.LoginEmailUserCommand,
)(dtos.LoginEmailUserResult, error){
	var session *aggregates.Session
	var tokenPair services.TokenPair

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		metadata := services.SessionMetadata{
			IPAddress: cmd.IPAddress,
			UserAgent: cmd.UserAgent,
			DeviceID: cmd.DeviceID,
		}
	

		var err error
		tokenPair, err = h.tokenService.Generate(
			ctx, 
			user.ID(),
			user.User.Role.String(),
			user.User.Email.String()
			metadata,
		)

		if err != nil{
			return fmt.Errorf("failed to generate tokens: %w", err)
		}

		sessionEntity := user.Login(
			cmd.IPAddress,
			cmd.UserAgent,
			cmd.DeviceID,
			tokenPair.RefreshToken,
			tokenPair.AccessToken,
			tokenPair.ExpiresAt,
		)

		if err := h.sessionRepo.Create(ctx, sessionEntity); err != nil{
			return fmt.Errorf("failed to create session: %w", err)
		}

		if err := h.userRepo.Update(ctx, user); err != nil{
			return fmt.Errorf("failed to update user last login: %w", err)
		}

		session = sessionEntity
		return nil
	})

	if err != nil {
		_ = h.publishFailedLoginEvent(ctx, cmd, err)
		return dtos.LoginEmailUserResult{}, err
	}	

	if err := h.publishSuccessLoginEvent(ctx, user, cmd); err != nil{
		h.logger.Error(ctx, "Failed to publish login success event",)
			zap.Error(err),
			zap.String("user_id", user.ID()),
	}

	h.logger.Info(ctx, "Email user logged in successfully",
		zap.String("user_id", user.ID()),
		zap.String("email", cmd.Email),
		zap.String("session_id", session.ID()),
	)

	return dtos.LoginEmailUserResult{
		UserID:       user.ID(),
		Email:        user.User.Email.String(),
		Username:     user.User.Username.String(),
		FirstName:    user.User.FirstName,
		LastName:     user.User.LastName,
		Role:         user.User.Role.String(),
		RequiresOTP: false,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Unix(),
		ExpiresIn:    int64(tokenPair.ExpiresAt.Sub(services.Now()).Seconds()),
		SessionID:    session.ID(),
	}, nil
}

func (h *LoginEmailHandler) publishSuccessLoginEvent(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.LoginEmailUserCommand,
) error{
	event := domain.NewUserLoginEvent(
		user.ID(),
		user.User.Email.String(),
		cmd.IPAddress,
		cmd.UserAgent,
		cmd.DeviceID,
		"SUCCESS",
		"email"
	)

	outboxMsg := &persistence.OutboxMessage{
		ID:          event.EventID().String(),
		EventType:   event.EventName(),
		AggregateID: event.AggregateID(),
		Payload:     event.Payload(),
		Metadata:    event.Metadata(),
		OccurredAt:  event.OccurredAt().Unix(),
	}
	
	if err := h.outbox.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

func (h *LoginEmailHandler) publishFailedLoginEvent(
	ctx context.Context, 
	cmd commands.LoginEmailUserCommand,
	cause error,
) error{
	event := domain.NewUserLoginEvent(
		"nil",
		cmd.Email,
		cmd.IPAddress,
		cmd.UserAgent,
		cmd.DeviceID,
		"FAILURE",
		"email",
	)

	meta := event.Metadata()
	meta["error"] = cause.Error()

	outboxMsg := &persistence.OutboxMessage{
		ID:          event.EventID().String(),
		EventType:   event.EventName(),
		AggregateID: event.AggregateID(),
		Payload:     event.Payload(),
		Metadata:    event.Metadata(),
		OccurredAt:  event.OccurredAt().Unix(),
	}

	if err := h.outbox.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}