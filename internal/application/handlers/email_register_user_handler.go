package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/services"
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

	registerCmd, ok := cmd.(commands.RegisterEmailCommand){
		if !ok{
		err := fmt.Errorf("invalid command type: expected RegisterEmailUserCommand, got %T", cmd)
span.RecordError()
span.SetStatus(codes.Error, "invalid command type" got %T,  )		}
	}
}
