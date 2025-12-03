package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"authentication/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type RegisterEmailHandler struct {
	userRepo       repositories.UserRepository
	auditRepo      repositories.AuditRepository
	uow            persistence.UnitOfWork
	passwordHasher *services.PasswordHashingService
	logger         logging.Logger
	tracer         tracing.Tracer
}

func NewRegisterEmailHandler(
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditRepository,
	uow persistence.UnitOfWork,
	passwordHasher *services.PasswordHashingService,
	logger logging.Logger,
	tracer tracing.Tracer,
) messaging.CommandHandler {
	return &RegisterEmailHandler{
		userRepo:       userRepo,
		auditRepo:      auditRepo,
		uow:            uow,
		passwordHasher: passwordHasher,
		logger:         logger.With(zap.String("handler", "register_email")),
		tracer:         tracer,
	}
}

func (r *RegisterEmailHandler) Handle(ctx context.Context, cmd messaging.Command) error {
	ctx, span := r.tracer.StartSpan(ctx, "RegisterEmailHandler.Handle",
		trace.WithSpanKind(trace.SpanKindInternal))

	defer span.End()

	registerCmd, ok := cmd.(commands.RegisterEmailUserCommand)
	if !ok {
		err := fmt.Errorf("invalid command type: expected RegisterEmailUserCommand, got %T", cmd)
		span.RecordError(err)
		//span.SetStatus(codes.Error, err.Error())
		return err
	}

	r.logger.Info(ctx, "Processing email user registration",
		zap.String("email", registerCmd.Email),
		zap.String("username", registerCmd.Username))

	defer func() {
		if h := recover(); h != nil {
			recErr := fmt.Errorf("panic: %v", h)
			r.logger.Error(ctx, "Panic during email user registration",
				zap.Any("panic", h),
				zap.String("email", registerCmd.Email),
			)
			r.logFailedRegistration(ctx, registerCmd, recErr)
			panic(r)
		}
	}()

	var user *aggregates.UserAggregate
	err := r.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		user, err = r.registerEmailUser(ctx, registerCmd)
		if err != nil {
			return err
		}
		//return
		//return h.outboxService.StoreEvents(ctx, user.GetEvents())
		return nil

	})

	if err != nil {
		r.logger.Error(ctx, "Email registration failed",
			zap.String("email", registerCmd.Email),
			zap.Error(err),
		)

		r.logFailedRegistration(ctx, registerCmd, err)
		span.RecordError(err)
		span.SetAttributes(attribute.String("registeration.status", "failed"))
		return err
	}

	r.logSuccessfulRegistration(ctx, user, registerCmd)

	r.logger.Info(ctx, "Email user registered successfully",
		zap.String("user_id", user.ID()),
		zap.String("email", registerCmd.Email),
		zap.String("username", registerCmd.Username),
	)

	span.SetAttributes(
		attribute.String("user.id", user.ID()),
		attribute.String("registration.type", "email"),
	)
	//span.SetStatus(codes.Ok, "success")

	return nil

}

func (r *RegisterEmailHandler) registerEmailUser(
	ctx context.Context,
	cmd commands.RegisterEmailUserCommand,
) (*aggregates.UserAggregate, error) {

	emailVO, err := valueobjects.NewEmail(cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	exists, err := r.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if exists {
		//call login
		return nil, domain.ErrEmailAlreadyInUse
	}

	usernameVO, err := valueobjects.NewUsername(cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	hashedPassword, err := r.passwordHasher.HashPassword(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role, err := valueobjects.NewRole(cmd.Role)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	user := aggregates.NewEmailUserAggregate(
		usernameVO,
		emailVO,
		hashedPassword,
		cmd.FirstName,
		cmd.LastName,
		role,
	)

	if err := r.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *RegisterEmailHandler) logSuccessfulRegistration(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.RegisterEmailUserCommand,
) {
	_ = r.auditRepo.Create(ctx, &aggregates.AuditLog{
		UserID:       user.User.ID,
		ResourceID:   user.User.ID,
		ResourceType: "user",
		Action:       valueobjects.AuditActionUserRegistered,
		Status:       "SUCCESS",
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Metadata: map[string]any{
			"type":     "email",
			"username": user.User.Username.String(),
			"email":    user.User.Email.String(),
			"role":     cmd.Role,
		},
	})
}

func (r *RegisterEmailHandler) logFailedRegistration(
	ctx context.Context,
	cmd commands.RegisterEmailUserCommand,
	err error,
) {
	_ = r.auditRepo.Create(ctx, &aggregates.AuditLog{
		Action:       valueobjects.AuditActionUserRegistered,
		Status:       "FAILURE",
		ResourceType: "user",
		ErrorMessage: err.Error(),
		IPAddress:    cmd.IPAddress,
		UserAgent:    cmd.UserAgent,
		Metadata: map[string]any{
			"email":    cmd.Email,
			"username": cmd.Username,
			"type":     "email",
		},
	})
}
