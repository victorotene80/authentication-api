package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/contracts/services"
	appDtos "authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	domainServices "authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type LoginEmailUserHandler struct {
	userRepo       repositories.UserRepository
	sessionRepo    repositories.SessionRepository
	auditRepo      repositories.AuditRepository
	uow            persistence.UnitOfWork
	//outboxService  *OutboxService
	passwordHasher *domainServices.PasswordHashingService
	tokenService   services.TokenService
	logger         logging.Logger
	tracer         tracing.Tracer
}

func NewLoginEmailUserHandler(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	auditRepo repositories.AuditRepository,
	uow persistence.UnitOfWork,
	//outboxService *OutboxService,
	passwordHasher *domainServices.PasswordHashingService,
	tokenService services.TokenService,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.CommandHandler {
	return &LoginEmailUserHandler{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		auditRepo:      auditRepo,
		uow:            uow,
		//outboxService:  outboxService,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
		logger:         logger.With(zap.String("handler", "login_email_user")),
		tracer:         tracer,
	}
}

func (h *LoginEmailUserHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	ctx, span := h.tracer.StartSpan(ctx, "LoginEmailUserHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	loginCmd, ok := cmd.(commands.LoginEmailUserCommand)
	if !ok {
		err := fmt.Errorf("invalid command type: expected LoginEmailUserCommand, got %T", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid command type")
		return err
	}

	h.logger.Info(ctx, "Processing email login",
		zap.String("email", loginCmd.Email),
	)

	defer func() {
		if r := recover(); r != nil {
			recErr := fmt.Errorf("panic: %v", r)
			h.logger.Error(ctx, "Panic during email login",
				zap.Any("panic", r),
				zap.String("email", loginCmd.Email),
			)
			h.logFailedLogin(ctx, loginCmd.Email, loginCmd.IPAddress, loginCmd.UserAgent, recErr)
			//panic(r)
		}
	}()

	var result *appDtos.LoginResult
	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = h.loginEmailUser(ctx, loginCmd)
		return err
	})

	if err != nil {
		h.logger.Error(ctx, "Email login failed",
			zap.Error(err),
			zap.String("email", loginCmd.Email),
		)
		h.logFailedLogin(ctx, loginCmd.Email, loginCmd.IPAddress, loginCmd.UserAgent, err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("login.status", "failed"))
		return err
	}

	h.logSuccessfulLogin(ctx, result.UserID, loginCmd.Email, loginCmd.IPAddress, loginCmd.UserAgent)

	h.logger.Info(ctx, "Email user logged in successfully",
		zap.String("user_id", result.UserID),
		zap.String("email", loginCmd.Email),
	)

	span.SetAttributes(
		attribute.String("user.id", result.UserID),
		attribute.String("login.type", "email"),
	)
	span.SetStatus(codes.Ok, "success")

	// Store result in context for HTTP handler to retrieve
	ctx = context.WithValue(ctx, "login_result", result)

	return nil
}

func (h *LoginEmailUserHandler) loginEmailUser(
	ctx context.Context,
	cmd commands.LoginEmailUserCommand,
) (*appDtos.LoginResult, error) {
	email, err := valueobjects.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.User.IsOAuthUser() {
		//send OTP to email user logs in with the OTP generated 
		return nil, fmt.Errorf("this account uses OAuth login with %s", user.User.Provider())
	}

	if !user.User.IsActive {
		return nil, domain.ErrUserLocked
	}

	if !h.passwordHasher.Verify(cmd.Password, user.User.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.User.IsVerified {
		return nil, fmt.Errorf("email not verified")
	}

	tokenClaims := services.TokenClaims{
		UserID:   user.ID(),
		Email:    user.User.Email.String(),
		Username: user.User.Username.String(),
		Role:     user.User.Role.String(),
	}

	tokens, err := h.tokenService.Generate(ctx, tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	session := user.Login(
		cmd.IPAddress,
		cmd.UserAgent,
		tokens.RefreshToken,
		tokens.AccessToken,
		tokens.ExpiresAt,
	)

	if err := h.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	user.User.RecordLogin()

	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Store events in outbox
	//if err := h.outboxService.StoreEvents(ctx, user.GetEvents()); err != nil {
	//	return nil, fmt.Errorf("failed to store events: %w", err)
	//}

	return &appDtos.LoginResult{
		UserID:       user.ID(),
		Email:        user.User.Email.String(),
		Username:     user.User.Username.String(),
		FirstName:    user.User.FirstName,
		LastName:     user.User.LastName,
		Role:         user.User.Role.String(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Unix(),
		IsNewUser:    false,
	}, nil
}

func (h *LoginEmailUserHandler) logSuccessfulLogin(
	ctx context.Context,
	userID, email, ipAddress, userAgent string,
) {
	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		UserID:       userID,
		ResourceID:   userID,
		ResourceType: "user",
		Action:       valueobjects.AuditActionUserLoggedIn,
		Status:       "SUCCESS",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata: map[string]any{
			"email":      email,
			"login_type": "email",
		},
	})
}

func (h *LoginEmailUserHandler) logFailedLogin(
	ctx context.Context,
	email, ipAddress, userAgent string,
	err error,
) {
	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		Action:       valueobjects.AuditActionUserLoggedIn,
		Status:       "FAILURE",
		ResourceType: "user",
		ErrorMessage: err.Error(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata: map[string]any{
			"email":      email,
			"login_type": "email",
		},
	})
}