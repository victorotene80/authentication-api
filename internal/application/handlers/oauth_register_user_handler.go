package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	oauthService "authentication/internal/application/contracts/services"
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

type RegisterOAuthUserHandler struct {
	userRepo      repositories.UserRepository
	auditRepo     repositories.AuditRepository
	uow           persistence.UnitOfWork
	//outboxService *OutboxService
	oauthService  oauthService.OAuthService
	logger        logging.Logger
	tracer        tracing.Tracer
}

func NewRegisterOAuthUserHandler(
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditRepository,
	uow persistence.UnitOfWork,
	//outboxService *OutboxService,
	oauthService oauthService.OAuthService,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.CommandHandler {
	return &RegisterOAuthUserHandler{
		userRepo:      userRepo,
		auditRepo:     auditRepo,
		uow:           uow,
		//outboxService: outboxService,
		oauthService:  oauthService,
		logger:        logger.With(zap.String("handler", "register_oauth_user")),
		tracer:        tracer,
	}
}

func (h *RegisterOAuthUserHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	ctx, span := h.tracer.StartSpan(ctx, "RegisterOAuthUserHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	registerCmd, ok := cmd.(commands.RegisterOAuthUserCommand)
	if !ok {
		err := fmt.Errorf("invalid command type: expected RegisterOAuthUserCommand, got %T", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid command type")
		return err
	}

	h.logger.Info(ctx, "Processing OAuth user registration",
		zap.String("provider", registerCmd.OAuthProvider),
	)

	defer func() {
		if r := recover(); r != nil {
			recErr := fmt.Errorf("panic: %v", r)
			h.logger.Error(ctx, "Panic during OAuth user registration",
				zap.Any("panic", r),
				zap.String("provider", registerCmd.OAuthProvider),
			)
			h.logFailedRegistration(ctx, registerCmd, recErr)
			panic(r) // Re-panic after logging
		}
	}()

	var user *aggregates.UserAggregate
	var isNewUser bool

	err := h.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		user, isNewUser, err = h.registerOAuthUser(ctx, registerCmd)
		if err != nil {
			return err
		}

		// Only store events for new users
		if isNewUser {
			//return h.outboxService.StoreEvents(ctx, user.GetEvents())
		}
		return nil
	})

	if err != nil {
		h.logger.Error(ctx, "Failed to register OAuth user",
			zap.Error(err),
			zap.String("provider", registerCmd.OAuthProvider),
		)
		h.logFailedRegistration(ctx, registerCmd, err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("registration.status", "failed"))
		return err
	}

	if isNewUser {
		h.logSuccessfulRegistration(ctx, user, registerCmd)
		h.logger.Info(ctx, "New OAuth user registered successfully",
			zap.String("user_id", user.ID()),
			zap.String("provider", registerCmd.OAuthProvider),
		)
	} else {
		h.logger.Info(ctx, "Existing OAuth user returned (idempotent)",
			zap.String("user_id", user.ID()),
			zap.String("provider", registerCmd.OAuthProvider),
		)
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID()),
		attribute.String("registration.type", "oauth"),
		attribute.Bool("is_new_user", isNewUser),
	)
	span.SetStatus(codes.Ok, "success")

	return nil
}

func (h *RegisterOAuthUserHandler) registerOAuthUser(
	ctx context.Context,
	cmd commands.RegisterOAuthUserCommand,
) (*aggregates.UserAggregate, bool, error) {
	ctx, span := h.tracer.StartSpan(ctx, "RegisterOAuthUserHandler.registerOAuthUser")
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
		return nil, false, fmt.Errorf("oauth verification failed: %w", err)
	}

	emailVO, err := valueobjects.NewEmail(info.Email)
	if err != nil {
		return nil, false, fmt.Errorf("invalid email from oauth provider: %w", err)
	}

	if !info.EmailVerified {
		return nil, false, fmt.Errorf("oauth email is not verified")
	}

	existingUser, err := h.userRepo.FindByEmail(ctx, emailVO)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, false, fmt.Errorf("failed to check email existence: %w", err)
	}

	if existingUser != nil {
		if existingUser.User.IsOAuthUser() && existingUser.User.Provider() == cmd.OAuthProvider {
			h.logger.Info(ctx, "OAuth user already exists, returning existing user",
				zap.String("email", info.Email),
				zap.String("provider", cmd.OAuthProvider),
			)
			return existingUser, false, nil // false = not a new user
		}

		//if user already exists login 
		// User registered via email or different OAuth provider
		return nil, false, domain.ErrEmailAlreadyInUse
	}

	// User doesn't exist - create new one
	role, err := valueobjects.NewRole(cmd.Role)
	if err != nil {
		return nil, false, fmt.Errorf("invalid role: %w", err)
	}

	user := aggregates.NewOAuthUserAggregate(
		emailVO,
		info.FirstName,
		info.LastName,
		cmd.OAuthProvider,
		role,
	)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID()),
		attribute.String("registration.type", "oauth"),
	)

	return user, true, nil // true = new user created
}

func (h *RegisterOAuthUserHandler) logSuccessfulRegistration(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.RegisterOAuthUserCommand,
) {
	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		UserID:       user.User.ID,
		ResourceID:   user.User.ID,
		ResourceType: "user",
		Action:       valueobjects.AuditActionUserRegistered,
		Status:       "SUCCESS",
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Metadata: map[string]any{
			"type":           "oauth",
			"oauth_provider": cmd.OAuthProvider,
			"email":          user.User.Email.String(),
			"role":           cmd.Role,
		},
	})
}

func (h *RegisterOAuthUserHandler) logFailedRegistration(
	ctx context.Context,
	cmd commands.RegisterOAuthUserCommand,
	err error,
) {
	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		Action:       valueobjects.AuditActionUserRegistered,
		Status:       "FAILURE",
		ResourceType: "user",
		ErrorMessage: err.Error(),
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Metadata: map[string]any{
			"type":           "oauth",
			"oauth_provider": cmd.OAuthProvider,
		},
	})
}