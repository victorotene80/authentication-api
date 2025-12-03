package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/contracts/services"
	oauthService "authentication/internal/application/contracts/services"
	appDtos "authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type LoginOAuthUserHandler struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	auditRepo   repositories.AuditRepository
	uow         persistence.UnitOfWork
	//outboxService *OutboxService
	oauthService oauthService.OAuthService
	tokenService services.TokenService
	logger       logging.Logger
	tracer       tracing.Tracer
}

func NewLoginOAuthUserHandler(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	auditRepo repositories.AuditRepository,
	uow persistence.UnitOfWork,
	//outboxService *OutboxService,
	oauthService oauthService.OAuthService,
	tokenService services.TokenService,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.CommandHandler {
	return &LoginOAuthUserHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		uow:         uow,
		//outboxService: outboxService,
		oauthService: oauthService,
		tokenService: tokenService,
		logger:       logger.With(zap.String("handler", "login_oauth_user")),
		tracer:       tracer,
	}
}

func (h *LoginOAuthUserHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	ctx, span := h.tracer.StartSpan(ctx, "LoginOAuthUserHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	loginCmd, ok := cmd.(commands.LoginOAuthUserCommand)
	if !ok {
		err := fmt.Errorf("invalid command type: expected LoginOAuthUserCommand, got %T", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid command type")
		return err
	}

	h.logger.Info(ctx, "Processing OAuth login",
		zap.String("provider", loginCmd.OAuthProvider),
	)

	defer func() {
		if r := recover(); r != nil {
			recErr := fmt.Errorf("panic: %v", r)
			h.logger.Error(ctx, "Panic during OAuth login",
				zap.Any("panic", r),
				zap.String("provider", loginCmd.OAuthProvider),
			)
			h.logFailedLogin(ctx, "", loginCmd.OAuthProvider, loginCmd.IPAddress, loginCmd.UserAgent, recErr)
			//panic(r)
		}
	}()

	var result *appDtos.LoginResult
	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		result, err = h.loginOAuthUser(ctx, loginCmd)
		return err
	})

	if err != nil {
		h.logger.Error(ctx, "OAuth login failed",
			zap.Error(err),
			zap.String("provider", loginCmd.OAuthProvider),
		)
		h.logFailedLogin(ctx, "", loginCmd.OAuthProvider, loginCmd.IPAddress, loginCmd.UserAgent, err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("login.status", "failed"))
		return err
	}

	h.logSuccessfulLogin(ctx, result.UserID, result.Email, loginCmd.OAuthProvider, loginCmd.IPAddress, loginCmd.UserAgent, result.IsNewUser)

	if result.IsNewUser {
		h.logger.Info(ctx, "New OAuth user registered and logged in",
			zap.String("user_id", result.UserID),
			zap.String("provider", loginCmd.OAuthProvider),
		)
	} else {
		h.logger.Info(ctx, "Existing OAuth user logged in successfully",
			zap.String("user_id", result.UserID),
			zap.String("provider", loginCmd.OAuthProvider),
		)
	}

	span.SetAttributes(
		attribute.String("user.id", result.UserID),
		attribute.String("login.type", "oauth"),
		attribute.Bool("is_new_user", result.IsNewUser),
	)
	span.SetStatus(codes.Ok, "success")

	// Store result in context for HTTP handler to retrieve
	ctx = context.WithValue(ctx, "login_result", result)

	return nil
}

func (h *LoginOAuthUserHandler) loginOAuthUser(
	ctx context.Context,
	cmd commands.LoginOAuthUserCommand,
) (*appDtos.LoginResult, error) {
	ctx, span := h.tracer.StartSpan(ctx, "LoginOAuthUserHandler.loginOAuthUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("oauth.provider", cmd.OAuthProvider),
	)

	info, err := h.oauthService.Verify(
		ctx,
		cmd.OAuthProvider,
		cmd.IDToken,
		cmd.AccessToken,
	)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("oauth verification failed: %w", err)
	}

	emailVO, err := valueobjects.NewEmail(info.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email from oauth provider: %w", err)
	}

	if !info.EmailVerified {
		return nil, fmt.Errorf("oauth email is not verified")
	}

	user, err := h.userRepo.FindByEmail(ctx, emailVO)
	isNewUser := false

	//is it not better to call the registeration handler than using "createOAuthUser"
	if err == domain.ErrUserNotFound {
		user, err = h.createOAuthUser(ctx, info, cmd.OAuthProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to create oauth user: %w", err)
		}
		isNewUser = true
	} else if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	//revalidate this
	// If user exists but registered via email/password, prevent OAuth login
	//send OTP
	if !user.User.IsOAuthUser() {
		return nil, fmt.Errorf("this email is registered with password login")
	}

	// If user exists but with different OAuth provider
	if user.User.Provider() != cmd.OAuthProvider {
		return nil, fmt.Errorf("this email is registered with %s", user.User.Provider())
	}

	if !user.User.IsActive {
		return nil, domain.ErrUserLocked
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

	// Persist session
	if err := h.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	user.User.RecordLogin()

	// Update user (last login time)
	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Store events in outbox
	//if err := h.outboxService.StoreEvents(ctx, user.GetEvents()); err != nil {
	//	return nil, fmt.Errorf("failed to store events: %w", err)
	//}

	span.SetAttributes(
		attribute.String("user.id", user.ID()),
		attribute.Bool("is_new_user", isNewUser),
	)

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
		IsNewUser:    isNewUser,
	}, nil
}

func (h *LoginOAuthUserHandler) createOAuthUser(
	ctx context.Context,
	info *oauthService.OAuthUserInfo,
	provider string,
) (*aggregates.UserAggregate, error) {
	emailVO, err := valueobjects.NewEmail(info.Email)
	if err != nil {
		return nil, err
	}

	role, err := valueobjects.NewRole("user")
	if err != nil {
		return nil, err
	}

	user := aggregates.NewOAuthUserAggregate(
		emailVO,
		info.FirstName,
		info.LastName,
		provider,
		role,
	)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	h.logger.Info(ctx, "Auto-registered new OAuth user",
		zap.String("user_id", user.ID()),
		zap.String("email", info.Email),
		zap.String("provider", provider),
	)

	return user, nil
}

func (h *LoginOAuthUserHandler) logSuccessfulLogin(
	ctx context.Context,
	userID, email, provider, ipAddress, userAgent string,
	isNewUser bool,
) {
	metadata := map[string]any{
		"email":          email,
		"login_type":     "oauth",
		"oauth_provider": provider,
		"is_new_user":    isNewUser,
	}

	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		UserID:       userID,
		ResourceID:   userID,
		ResourceType: "user",
		Action:       valueobjects.AuditActionUserLoggedIn,
		Status:       "SUCCESS",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata:     metadata,
	})
}

func (h *LoginOAuthUserHandler) logFailedLogin(
	ctx context.Context,
	email, provider, ipAddress, userAgent string,
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
			"email":          email,
			"login_type":     "oauth",
			"oauth_provider": provider,
		},
	})
}
