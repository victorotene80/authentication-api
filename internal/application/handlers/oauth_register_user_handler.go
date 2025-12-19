package handlers

import (
	"context"
	"fmt"

	"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/internal/application/contracts/persistence"
	"authentication/internal/application/contracts/services"
	"authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/repositories"
	"authentication/internal/domain/valueobjects"
	"authentication/internal/domain/events"
	"authentication/shared/logging"

	"go.uber.org/zap"
)

type RegisterOAuthUserHandler struct {
	userRepo     repositories.UserRepository
	auditRepo    repositories.AuditRepository
	outboxRepo   persistence.OutboxRepository
	uow          persistence.UnitOfWork
	oauthService services.OAuthService
	tokenService services.TokenService
	logger       logging.Logger
}

func NewRegisterOAuthUserHandler(
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditRepository,
	outboxRepo persistence.OutboxRepository,
	uow persistence.UnitOfWork,
	oauthService services.OAuthService,
	tokenService services.TokenService,
	logger logging.Logger,
) messaging.CommandHandler[commands.RegisterOAuthUserCommand, dtos.RegisterOAuthUserResult] {
	return &RegisterOAuthUserHandler{
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		outboxRepo:   outboxRepo,
		uow:          uow,
		oauthService: oauthService,
		tokenService: tokenService,
		logger:       logger.With(zap.String("handler", "register_oauth_user")),
	}
}

func (r *RegisterOAuthUserHandler) Handle(
	ctx context.Context,
	cmd commands.RegisterOAuthUserCommand,
) (dtos.RegisterOAuthUserResult, error) {
	defer func() {
		if h := recover(); h != nil {
			r.logger.Error(ctx, "Panic during OAuth user registration",
				zap.Any("panic", h),
				zap.String("email", cmd.Email),
			)

			_ = r.publishFailedUserCreatedEvent(ctx, cmd, fmt.Errorf("panic: %v", h))
			panic(h)
		}
	}()

	var user *aggregates.UserAggregate
	var isNewUser bool

	err := r.uow.Execute(ctx, func(ctx context.Context) error {
		var err error
		user, isNewUser, err = r.registerOAuthUser(ctx, cmd)
		return err
	})

	if err != nil {
		_ = r.publishFailedUserCreatedEvent(ctx, cmd, err)
		return dtos.RegisterOAuthUserResult{}, err
	}

	// Generate tokens for OAuth users
	metadata := services.SessionMetadata{
		IPAddress: cmd.IPAddress,
		UserAgent: cmd.UserAgent,
		DeviceID:  cmd.DeviceID,
	}

	tokenPair, err := r.tokenService.Generate(
		ctx,
		user.ID(),
		cmd.Role,
		user.User.Email.String(),
		metadata,
	)
	if err != nil {
		return dtos.RegisterOAuthUserResult{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if isNewUser {
		_ = r.publishSuccessUserCreatedEvent(ctx, user, cmd)
	}

	result := dtos.RegisterOAuthUserResult{
		UserID:          user.ID(),
		Email:           user.User.Email.String(),
		Username:        "",
		FirstName:       user.User.FirstName,
		LastName:        user.User.LastName,
		Role:            cmd.Role,
		IsOAuthUser:     true,
		OAuthProvider:   cmd.OAuthProvider,
		RequiresOnboard: isNewUser,
		IsNewUser:       isNewUser,
		AccessToken:     tokenPair.AccessToken,
		RefreshToken:    tokenPair.RefreshToken,
		ExpiresAt:       tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresIn:       tokenPair.ExpiresIn,
		SessionID:       tokenPair.SessionID,
	}

	return result, nil
}

func (h *RegisterOAuthUserHandler) registerOAuthUser(
	ctx context.Context,
	cmd commands.RegisterOAuthUserCommand,
) (*aggregates.UserAggregate, bool, error) {
	info, err := h.oauthService.Verify(
		ctx,
		cmd.OAuthProvider,
		cmd.IDToken,
		cmd.AccessToken,
	)
	if err != nil {
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

	/*if existingUser != nil {
		if existingUser.User.IsOAuthUser() && existingUser.User.Provider() == cmd.OAuthProvider {
			return existingUser, false, nil
		}
		return nil, false, domain.ErrEmailAlreadyInUse
	}*/

	if existingUser != nil{
		return existingUser, false, nil
	}

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

	return user, true, nil
}

func (r *RegisterOAuthUserHandler) publishSuccessUserCreatedEvent(
	ctx context.Context,
	user *aggregates.UserAggregate,
	cmd commands.RegisterOAuthUserCommand,
) error {
	event := events.NewUserCreatedEvent(
		user.ID(),
		"nil",//user.User.Username.String(),
		user.User.Email.String(),
		user.User.Role.String(),
		user.User.FirstName,
		user.User.LastName,
		cmd.IPAddress,
		cmd.UserAgent,
		"SUCCESS",
		"oauth",
	)

	outboxMsg := &persistence.OutboxMessage{
		ID:          event.EventID().String(),
		EventType:   event.EventName(),
		AggregateID: event.AggregateID(),
		Payload:     event.Payload(),
		Metadata:    event.Metadata(),
		OccurredAt:  event.OccurredAt().Unix(),
	}

	if err := r.outboxRepo.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

func (r *RegisterOAuthUserHandler) publishFailedUserCreatedEvent(
	ctx context.Context,
	cmd commands.RegisterOAuthUserCommand,
	cause error,
) error{
	event := events.NewUserCreatedEvent(
		"nil",
		"nil",
		cmd.Email,
		cmd.Role,
		"nil",
		"nil",
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

	if err := r.outboxRepo.Save(ctx, outboxMsg); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}