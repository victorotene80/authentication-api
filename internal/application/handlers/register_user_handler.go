package handlers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"authentication/internal/application/commands"
	"authentication/internal/domain"
	"authentication/internal/domain/entities"
	events"authentication/internal/application/contracts"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/infrastructure/observability/metrics"
	"authentication/internal/infrastructure/observability/tracing"
	"authentication/shared"
	"authentication/shared/logging"
	"authentication/shared/utils"
)

type RegisterUserHandler struct {
	userRepo   repositories.UserRepository
	outboxRepo repositories.OutboxRepository
	unitOfWork shared.UnitOfWork
	serializer events.EventSerializer
	logger     logging.Logger
}

func NewRegisterUserHandler(
	userRepo repositories.UserRepository,
	outboxRepo repositories.OutboxRepository,
	uow shared.UnitOfWork,
	serializer events.EventSerializer,
	logger logging.Logger,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:   userRepo,
		outboxRepo: outboxRepo,
		unitOfWork: uow,
		serializer: serializer,
		logger:     logger,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd commands.RegisterUserCommand) error {
	start := utils.NowUTC()
	log := h.logger.With(
		zap.String("command", "RegisterUserCommand"),
		zap.String("email", cmd.Email),
		zap.String("username", cmd.Username),
	)

	// Start distributed tracing span
	ctx, span := tracing.StartSpan(ctx, "RegisterUserHandler.Handle")
	defer span.End()

	tracing.AddSpanAttributes(span,
		tracing.CommandAttributes("RegisterUserCommand")...,
	)
	tracing.AddSpanAttributes(span,
		tracing.UserAttributes("", cmd.Email)...,
	)

	// Execute within transaction
	err := h.unitOfWork.Execute(ctx, func(txCtx context.Context) error {
		// Set repositories to use transaction
		h.userRepo.SetDB(h.unitOfWork.DB())
		h.outboxRepo.SetDB(h.unitOfWork.DB())

		// Validate and create value objects
		email, err := valueobjects.NewEmail(cmd.Email)
		if err != nil {
			log.Error(txCtx, "Invalid email format", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("invalid_email").Inc()
			metrics.ValidationErrors.WithLabelValues("email", "invalid_format").Inc()
			tracing.RecordError(span, err)
			return err
		}

		username, err := valueobjects.NewUsername(cmd.Username)
		if err != nil {
			log.Error(txCtx, "Invalid username format", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("invalid_username").Inc()
			metrics.ValidationErrors.WithLabelValues("username", "invalid_format").Inc()
			tracing.RecordError(span, err)
			return err
		}

		// Password validation and hashing
		passwordStart := time.Now()
		password, err := valueobjects.NewPassword(cmd.Password)
		passwordHashDuration := time.Since(passwordStart).Seconds()
		metrics.PasswordHashingDuration.Observe(passwordHashDuration)
		
		tracing.AddSpanEvent(span, "Password hashed",
			tracing.HTTPAttributes("", "", 0)...,
		)

		if err != nil {
			log.Error(txCtx, "Invalid password format", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("invalid_password").Inc()
			metrics.ValidationErrors.WithLabelValues("password", "weak_password").Inc()
			tracing.RecordError(span, err)
			return err
		}

		log.Debug(txCtx, "Value objects validated",
			zap.String("email", email.String()),
			zap.String("username", username.String()),
			zap.Float64("password_hash_duration_ms", passwordHashDuration*1000),
		)

		// Check for existing user by email
		existingUser, err := h.userRepo.FindByEmail(txCtx, email)
		if err != nil && !repositories.IsNotFoundError(err) {
			log.Error(txCtx, "Repository error checking email", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("repository_error").Inc()
			tracing.RecordError(span, err)
			return fmt.Errorf("failed to check email existence: %w", err)
		}

		if existingUser != nil {
			log.Warn(txCtx, "Email already registered")
			metrics.UserRegistrationErrors.WithLabelValues("email_already_exists").Inc()
			return domain.ErrEmailAlreadyInUse
		}

		// Check for existing username
		existingUsername, err := h.userRepo.FindByUsername(txCtx, username)
		if err != nil && !repositories.IsNotFoundError(err) {
			log.Error(txCtx, "Repository error checking username", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("repository_error").Inc()
			tracing.RecordError(span, err)
			return fmt.Errorf("failed to check username existence: %w", err)
		}

		if existingUsername != nil {
			log.Warn(txCtx, "Username already taken")
			metrics.UserRegistrationErrors.WithLabelValues("username_already_exists").Inc()
			return domain.ErrUsernameAlreadyTaken
		}

		tracing.AddSpanEvent(span, "Uniqueness checks passed")

		// Create new user entity
		user, err := entities.NewUser(email, username, password, cmd.FirstName, cmd.LastName)
		if err != nil {
			log.Error(txCtx, "Failed to create user entity", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("entity_creation_failed").Inc()
			tracing.RecordError(span, err)
			return err
		}

		tracing.AddSpanAttributes(span,
			tracing.UserAttributes(user.ID().String(), user.Email().String())...,
		)

		// Persist user
		if err := h.userRepo.Save(txCtx, user); err != nil {
			log.Error(txCtx, "Failed to persist user", zap.Error(err))
			metrics.UserRegistrationErrors.WithLabelValues("persistence_failed").Inc()
			tracing.RecordError(span, err)
			return fmt.Errorf("failed to save user: %w", err)
		}

		log.Info(txCtx, "User persisted successfully", zap.String("user_id", user.ID().String()))
		tracing.AddSpanEvent(span, "User persisted to database")

		// Store domain events in outbox
		domainEvents := user.DomainEvents()
		if len(domainEvents) > 0 {
			log.Debug(txCtx, "Processing domain events", zap.Int("event_count", len(domainEvents)))

			for _, e := range domainEvents {
				payload, err := h.serializer.Serialize(e)
				if err != nil {
					log.Error(txCtx, "Failed to serialize event", 
						zap.Error(err),
						zap.String("event_name", e.EventName()),
					)
					metrics.UserRegistrationErrors.WithLabelValues("event_serialization_failed").Inc()
					tracing.RecordError(span, err)
					return fmt.Errorf("failed to serialize event %s: %w", e.EventName(), err)
				}

				outboxEvent := entities.OutboxEvent{
					EventId:     e.EventID(),
					EventName:   e.EventName(),
					CreatedAt:   e.OccurredAt(),
					AggregateID: e.AggregateID(),
					Payload:     payload,
				}

				if err := h.outboxRepo.InsertTx(txCtx, h.unitOfWork.DB(), outboxEvent); err != nil {
					log.Error(txCtx, "Failed to insert outbox event",
						zap.Error(err),
						zap.String("event_name", e.EventName()),
					)
					metrics.UserRegistrationErrors.WithLabelValues("outbox_insert_failed").Inc()
					tracing.RecordError(span, err)
					return fmt.Errorf("failed to insert outbox event: %w", err)
				}

				// Record event metrics
				metrics.EventsPublished.WithLabelValues(e.EventName(), "pending").Inc()
				
				tracing.AddSpanEvent(span, "Event stored in outbox",
					tracing.EventAttributes(e.EventName(), e.AggregateID().)...,
				)
			}

			log.Info(txCtx, "All domain events stored in outbox",
				zap.Int("event_count", len(domainEvents)),
				zap.String("user_id", user.ID().String()),
			)
		}

		user.ClearEvents()
		return nil
	})

	duration := time.Since(start).Seconds()

	if err != nil {
		log.Error(ctx, "User registration failed",
			zap.Error(err),
			zap.Float64("duration_seconds", duration),
		)
		
		tracing.RecordError(span, err)
		
		// Record metrics
		metrics.CommandExecutionDuration.WithLabelValues("RegisterUserCommand").Observe(duration)
		metrics.CommandExecutionTotal.WithLabelValues("RegisterUserCommand", "error").Inc()
		
		return err
	}

	log.Info(ctx, "User registration completed successfully",
		zap.String("email", cmd.Email),
		zap.String("username", cmd.Username),
		zap.Float64("duration_seconds", duration),
	)

	// Record success metrics
	metrics.UserRegistrations.Inc()
	metrics.CommandExecutionDuration.WithLabelValues("RegisterUserCommand").Observe(duration)
	metrics.CommandExecutionTotal.WithLabelValues("RegisterUserCommand", "success").Inc()

	tracing.AddSpanEvent(span, "Registration completed successfully")

	return nil
}