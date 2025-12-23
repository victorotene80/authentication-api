package handlers

import (
	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/contracts/services"
	"authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type LoginOAuthHandler struct {
	userRepo     repositories.UserRepository
	sessionRepo  repositories.SessionRepository
	outbox       persistence.OutboxRepository
	uow          persistence.UnitOfWork
	oauthService services.OAuthService
	tokenService services.TokenService
	otpService   services.OTPService
	logger       logging.Logger
}

func NewLoginOAuthHandler(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	outbox persistence.OutboxRepository,
	uow persistence.UnitOfWork,
	oauthService services.OAuthService,
	tokenService services.TokenService,
	otpService services.OTPService,
	logger logging.Logger,
) {
	return &LoginOAuthHandler{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		outbox:       outbox,
		uow:          uow,
		oauthService: oauthService,
		tokenService: tokenService,
		otpService:   otpService,
		logger:       logger.With(zap.String("handler", "login_oauth")),
	}
}

func (h *LoginOAuthHandler) Handle(
	ctx context.Context,
	cmd commands.LoginOAuthUserCommand,
) (dtos.LoginOAuthUserResult, error) {

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "Panic during OAuth user login",
				zap.Any("panic", r),
				zap.String("email", cmd.Email),
			)

			_ = h.publishFailedUserLoginEvent(ctx, cmd, fmt.Errorf("panic: %v", r))
			panic(r)
		}
	}()

	info, err := h.oauthService.Verify(
		ctx,
		cmd.OAuthProvider,
		cmd.IDToken,
		cmd.AccessToken,
	)

	if err != nil {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("failed to verify oauth tokens: %w", err)
	}

	emailVO, err := valueobjects.NewEmail(info.Email)
	if err != nil {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("invalid email from oauth provider: %w", err)
	}

	if !info.EmailVerified {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("oauth email not verified")
	}

	existingUser, err := h.userRepo.FindByEmail(ctx, emailVO)
	if err != nil && err != domain.ErrUserNotFound {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("failed to check email existence: %w", err)
	}

	if existingUser == nil {
		return dtos.LoginOAuthUserResult{}, domain.ErrUserNotFound
	}

	if !existingUser.User.IsActive {
		_ = h.publishFailedLoginEvent(ctx, cmd, domain.ErrUserAlreadyInactive)
		return dtos.LoginOAuthUserResult{}, domain.ErrUserAlreadyInactive
	}

	//come back to this

	if !existingUser.User.IsOAuthUser() {
		return h.handleEmailRegisteredUserLogin(ctx, existingUser, cmd)
	}

	return h.generateTokensAndLogin(ctx, existingUser, cmd)

}

func (h *LoginOAuthHandler) handleEmailRegisteredUserLogin(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.LoginOAuthUserCommand,
) (dtos.LoginOAuthUserResult, error) {

	isLimited, err := h.otpService.IsRateLimited(ctx, cmd.Email)
	if err != nil {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("failed to check OTP rate limit: %w", err)
	}
	if isLimited {
		return dtos.LoginOAuthUserResult{}, domain.ErrOTPRateLimited
	}

	otpCode, err := h.otpService.Generate(ctx, cmd.Email, services.OTPPurposeLogin)
	if err != nil {
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := h.otpService.SendEmail(ctx, user.User.Email.String(), otpCode, services.OTPPurposeLogin); err != nil {
		h.logger.Error(ctx, "Failed to send login OTP email to OAuth registered user",
			zap.String("email", user.User.Email.String()),
			zap.Error(err),
		)
		return dtos.LoginOAuthUserResult{}, fmt.Errorf("failed to send OTP email: %w", err)
	}

	h.logger.Info(ctx, "OTP sent to email user attempting OAuth login",
		zap.String("email", user.User.Email.String()),
		zap.String("oauth_provider", cmd.OAuthProvider),
	)

	return dtos.LoginOAuthUserResult{
		Email:   user.User.Email.String(),
		RequiresOTP: true,
		OTPSent:	true,
		Message: fmt.Sprintf("This account uses email/password for sign-in. We've sent a verification code to %s.", user.User.Email.String()),
	}, nil
}

func (h *LoginOAuthHandler) generateTokensAndLogin(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.LoginOAuthUserCommand,
)(dtos.LoginOAuthUserResult, error){
	var session *aggregates.Session 
	var tokenPair *services.TokenPair
}