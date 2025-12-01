package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	oauthService "authentication/internal/application/contracts/services"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
)

type RegisterUserHandler struct {
	userRepo       repositories.UserRepository
	auditRepo      repositories.AuditRepository
	uow            persistence.UnitOfWork
	dispatcher     messaging.EventDispatcher
	passwordHasher *services.PasswordHashingService
	oauthService   oauthService.OAuthService
	logger         logging.Logger
	tracer         tracing.Tracer
}

func NewRegisterUserHandler(
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditRepository,
	uow persistence.UnitOfWork,
	dispatcher messaging.EventDispatcher,
	oauthService oauthService.OAuthService,
	passwordHasher *services.PasswordHashingService,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.CommandHandler {
	return &RegisterUserHandler{
		userRepo:       userRepo,
		auditRepo:      auditRepo,
		oauthService:   oauthService,
		uow:            uow,
		dispatcher:     dispatcher,
		passwordHasher: passwordHasher,
		logger:         logger.With(zap.String("handler", "register_user")),
		tracer:         tracer,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	ctx, span := h.tracer.StartSpan(ctx, "RegisterUserHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("command.type", fmt.Sprintf("%T", cmd))),
	)
	defer span.End()

	registerCmd, ok := cmd.(commands.RegisterUserCommand)
	if !ok {
		err := fmt.Errorf("invalid command type: expected RegisterUserCommand, got %T", cmd)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid command type")
		return err
	}

	/*h.logger.Info(ctx, "Processing user registration",
		zap.String("email", registerCmd.Email),
		zap.String("username", registerCmd.Username),
		zap.Bool("is_oauth", registerCmd.IsOAuthRegistration()),
		//zap.String("oauth_provider", registerCmd.OAuthProvider),
	)*/

	defer func() {
		if r := recover(); r != nil {
			recErr := fmt.Errorf("panic: %v", r)
			h.logger.Error(ctx, "Panic during user registration",
				zap.Any("panic", r),
				//zap.String("email", registerCmd.Email),
			)

			h.logFailedRegistration(
				ctx,
				registerCmd,
				recErr,
			)

			panic(r)
		}
	}()

	var user *aggregates.UserAggregate
	err := h.uow.Execute(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		user, err = h.registerUser(ctx, registerCmd)
		return err
	})

	if err != nil {
		h.logger.Error(ctx, "Failed to register user",
			zap.Error(err),
			zap.String("email", registerCmd.Email),
		)
		h.logFailedRegistration(
			ctx,
			registerCmd,
			err,
		)
		span.RecordError(err)
		span.SetAttributes(attribute.String("registration.status", "failed"))
		return err
	}

	if err := h.dispatcher.DispatchAll(ctx, user.GetEvents()); err != nil {
		h.logger.Error(ctx, "Failed to dispatch domain events",
			zap.Error(err),
			zap.String("user_id", user.ID()),
		)
	}

	h.logSuccessfulRegistration(
		ctx,
		user,
		registerCmd,
	)

	h.logger.Info(ctx, "User registered successfully",
		zap.String("user_id", user.ID()),
		zap.String("email", registerCmd.Email),
		zap.String("username", registerCmd.Username),
	)

	span.SetAttributes(attribute.String("user.id", user.ID()))
	span.SetStatus(codes.Ok, "success")

	return nil
}

func (h *RegisterUserHandler) registerUser(
	ctx context.Context,
	cmd commands.RegisterUserCommand,
) (*aggregates.UserAggregate, error) {
	//ctx, span := h.tracer.StartSpan(ctx, "RegisterUserHandler.registerUser")
	//defer span.End()

	if cmd.IsOAuthRegistration() {
		return h.registerOAuthUser(ctx, cmd)
	}

	email, err := valueobjects.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	exists, err := h.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyInUse
	}

	return h.registerEmailUser(ctx, cmd, email)

}

func (h *RegisterUserHandler) registerEmailUser(
	ctx context.Context,
	cmd commands.RegisterUserCommand,
	email valueobjects.Email,
) (*aggregates.UserAggregate, error) {

	username, err := valueobjects.NewUsername(cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	hashedPassword, err := h.passwordHasher.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role, err := valueobjects.NewRole(cmd.Role)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	user := aggregates.NewEmailUserAggregate(
		username,
		email,
		hashedPassword,
		//phone,
		cmd.FirstName,
		cmd.LastName,
		role,
	)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (h *RegisterUserHandler) registerOAuthUser(
	ctx context.Context,
	cmd commands.RegisterUserCommand,
	//email valueobjects.Email,
) (*aggregates.UserAggregate, error) {
	ctx, span := h.tracer.StartSpan(ctx, "RegisterUserHandler.registerOAuthUser")
	defer span.End()

	/*span.SetAttributes(
		attribute.String("oauth.provider", cmd.OAuthProvider),
		attribute.String("oauth.id", cmd.OAuthID),
	)*/

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

	exists, err := h.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyInUse
	}

	role, err := valueobjects.NewRole(cmd.Role)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	if !info.EmailVerified {
		return nil, fmt.Errorf("oauth email is not verified")
	}

	user := aggregates.NewOAuthUserAggregate(
		emailVO,
		info.FirstName,
		info.LastName,
		cmd.OAuthProvider,
		role,
	)

	if err := h.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID()),
		attribute.String("registration.type", "oauth"),
	)

	return user, nil
}

func getRegistrationType(cmd commands.RegisterUserCommand) string {
	if cmd.IsOAuthRegistration() {
		return "oauth"
	}
	return "email"
}

func (h *RegisterUserHandler) logSuccessfulRegistration(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.RegisterUserCommand,
) {
	_ = h.auditRepo.Create(ctx, &aggregates.AuditLog{
		UserID:       &user.User.ID,
		ResourceID:   &user.User.ID,
		ResourceType: "user",
		Action:       valueobjects.AuditActionUserRegistered,
		Status:       "SUCCESS",
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Metadata: map[string]any{
			"type":     getRegistrationType(cmd),
			"username": user.User.Username.String(),
			"role":     cmd.Role,
		},
	})
}

func (h *RegisterUserHandler) logFailedRegistration(
	ctx context.Context,
	cmd commands.RegisterUserCommand,
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
			"email":    cmd.Email,
			"username": cmd.Username,
			"type":     getRegistrationType(cmd),
		},
	})
}
