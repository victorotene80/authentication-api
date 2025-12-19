package handlers

import (
	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/events"
	"authentication/internal/domain/repositories"
	domainServices "authentication/internal/domain/services"
	"authentication/internal/domain/valueobjects"
	"authentication/shared/logging"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type RegisterEmailHandler struct {
	userRepo       repositories.UserRepository
	outbox         persistence.OutboxRepository
	uow            persistence.UnitOfWork
	passwordHasher *domainServices.PasswordHashingService
	logger         logging.Logger
}

func NewRegisterEmailHandler(
	userRepo repositories.UserRepository,
	outbox persistence.OutboxRepository,
	uow persistence.UnitOfWork,
	passwordHasher *domainServices.PasswordHashingService,
	logger logging.Logger,
) messaging.CommandHandler[commands.RegisterEmailUserCommand, dtos.RegisterEmailUserResult] {
	return &RegisterEmailHandler{
		userRepo:       userRepo,
		outbox:         outbox,
		uow:            uow,
		passwordHasher: passwordHasher,
		logger:         logger.With(zap.String("handler", "register_email")),
	}
}

func (r *RegisterEmailHandler) Handle(
	ctx context.Context,
	cmd commands.RegisterEmailUserCommand,
) (dtos.RegisterEmailUserResult, error) {
	defer func() {
		if h := recover(); h != nil {
			r.logger.Error(ctx, "Panic during email user registration",
				zap.Any("panic", h),
				zap.String("email", cmd.Email),
			)

			_ = r.publishFailedUserCreatedEvent(ctx, cmd, fmt.Errorf("panic: %v", h))
			panic(h)
		}
	}()

	var user *aggregates.UserAggregate
	err := r.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		user, err = r.registerEmailUser(ctx, cmd)
		return err
	})

	if err != nil {
		_ = r.publishFailedUserCreatedEvent(ctx, cmd, err)
		return dtos.RegisterEmailUserResult{}, err
	}

	if err := r.publishSuccessUserCreatedEvent(ctx, user, cmd); err != nil {
		r.logger.Error(ctx, "Failed to publish user created event after email registration",
			zap.Error(err),
			zap.String("user_id", user.ID()),
			zap.String("email", cmd.Email),
		)
	}

	result := dtos.RegisterEmailUserResult{
		UserID:          user.ID(),
		Email:           user.User.Email.String(),
		Username:        user.User.Username.String(),
		FirstName:       user.User.FirstName,
		LastName:        user.User.LastName,
		Role:            cmd.Role,
		IsOAuthUser:     false,
		OAuthProvider:   "",
		RequiresOnboard: false,
		IsNewUser:       true,
	}

	return result, nil
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

	r.logger.Info(ctx, "Email user registered successfully",
		zap.String("user_id", user.ID()),
		zap.String("email", cmd.Email),
		zap.String("username", cmd.Username),
	)

	return user, nil
}

func (r *RegisterEmailHandler) publishSuccessUserCreatedEvent(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.RegisterEmailUserCommand,
) error {
	event := events.NewUserCreatedEvent(
		user.ID(),
		user.User.Username.String(),
		user.User.Email.String(),
		user.User.Role.String(),
		user.User.FirstName,
		user.User.LastName,
		cmd.IPAddress,
		cmd.UserAgent,
		"SUCCESS",
		"email",
	)

	outboxMsg := &persistence.OutboxMessage{
		ID:          event.EventID().String(),
		EventType:   event.EventName(),
		AggregateID: event.AggregateID(),
		Payload:     event.Payload(),
		Metadata:    event.Metadata(),
		OccurredAt:  event.OccurredAt().Unix(),
	}

	if err := r.outbox.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

func (r *RegisterEmailHandler) publishFailedUserCreatedEvent(
	ctx context.Context,
	cmd commands.RegisterEmailUserCommand,
	cause error,
) error {

	event := events.NewUserCreatedEvent(
		"nil",
		cmd.Username,
		cmd.Email,
		cmd.Role,
		cmd.FirstName,
		cmd.LastName,
		cmd.IPAddress,
		cmd.UserAgent,
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

	if err := r.outbox.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil

}
