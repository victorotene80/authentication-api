package bootstrap

import (
	"authentication/api/handlers"
	"authentication/api/routes"
	"authentication/internal/application/commands"
	"authentication/internal/application/contracts"
	"authentication/internal/application/contracts/background"
	"authentication/internal/application/contracts/messaging"
	appHandlers "authentication/internal/application/handlers"
	"authentication/internal/domain/repositories"
	outboxProcessor "authentication/internal/infrastructure/background"
	bus "authentication/internal/infrastructure/messaging/bus"
	"authentication/internal/infrastructure/messaging/event"
	dbRepositories "authentication/internal/infrastructure/persistence/repositories"
	uw "authentication/internal/infrastructure/persistence/uow"
	"authentication/shared/config"
	"authentication/shared/logging"
	"authentication/shared/persistence"
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Container struct {
	Config          *config.Config
	DB              *sql.DB
	Logger          logging.Logger
	UserRepo        repositories.UserRepository
	OutboxRepo      contracts.OutboxRepository
	UnitOfWork      persistence.UnitOfWork
	EventSerializer messaging.EventSerializer
	EventDispatcher messaging.EventDispatcher
	CommandBus      messaging.CommandBus
	QueryBus        messaging.QueryBus
	OutboxProcessor background.OutboxProcessor
	httpServer      *http.Server
}

func NewContainer(ctx context.Context, cfg *config.Config, db *sql.DB) (*Container, error) {
	c := &Container{
		Config: cfg,
		DB:     db,
		Logger: logging.Get().With(
			zap.String("service", "authentication"),
			zap.String("environment", cfg.App.Environment),
		),
	}

	// Initialize dependencies
	c.initRepositories()
	c.initUnitOfWork()
	c.initEventSystem()
	if err := c.initCommandBus(); err != nil {
		return nil, fmt.Errorf("failed to initialize command bus: %w", err)
	}
	if err := c.initQueryBus(); err != nil {
		return nil, fmt.Errorf("failed to initialize query bus: %w", err)
	}
	c.initOutboxProcessor()

	// Initialize HTTP server
	c.initHTTPServer()

	c.Logger.Info(ctx, "Dependency container fully initialized")
	return c, nil
}

// ---------------------- Initialization Helpers ----------------------

func (c *Container) initRepositories() {
	baseUserRepo := dbRepositories.NewPostgresUserRepository(c.DB)
	c.UserRepo = dbRepositories.NewInstrumentedUserRepository(baseUserRepo)
	c.OutboxRepo = dbRepositories.NewPostgresOutboxRepository(c.DB)

	c.Logger.Debug(context.Background(), "Repositories initialized")
}

func (c *Container) initUnitOfWork() {
	baseUoW := uw.NewUnitOfWork(c.DB)
	c.UnitOfWork = uw.NewInstrumentedUnitOfWork(baseUoW)

	c.Logger.Debug(context.Background(), "Unit of Work initialized")
}

func (c *Container) initEventSystem() {
	c.EventSerializer = event.NewJSONEventSerializer()
	c.EventDispatcher = event.NewCompositeEventDispatcher(false)

	c.Logger.Debug(context.Background(), "Event system initialized")
}

func (c *Container) initCommandBus() error {
	c.CommandBus = bus.NewCommandBus()
	handlers := map[messaging.Command]messaging.CommandHandler{
		commands.RegisterUserCommand{}: c.createRegisterUserHandler(),
	}
	for cmd, handler := range handlers {
		if err := c.CommandBus.Register(cmd, handler); err != nil {
			return fmt.Errorf("failed to register handler for %s: %w", cmd.CommandName(), err)
		}
		c.Logger.Debug(context.Background(), "Registered command handler", zap.String("command", cmd.CommandName()))
	}
	c.Logger.Info(context.Background(), "Command bus initialized", zap.Int("handlers", len(handlers)))
	return nil
}

func (c *Container) initQueryBus() error {
	c.QueryBus = bus.NewQueryBus()
	c.Logger.Info(context.Background(), "Query bus initialized")
	return nil
}

func (c *Container) initOutboxProcessor() {
	c.OutboxProcessor = outboxProcessor.NewOutboxProcessor(
		c.OutboxRepo,
		c.EventSerializer,
		c.Logger,
		c.EventDispatcher,
	)
	c.Logger.Debug(context.Background(), "Outbox processor initialized")
}

func (c *Container) initHTTPServer() {
	authHandler := handlers.NewAuthHandler(
		c.CommandBus,
	)

	router := routes.NewRouter(authHandler)

	c.httpServer = &http.Server{
		Addr:         c.Config.Server.GetServerAddr(), // <- use GetServerAddr()
		Handler:      router,
		ReadTimeout:  c.Config.Server.ReadTimeout,
		WriteTimeout: c.Config.Server.WriteTimeout,
		IdleTimeout:  c.Config.Server.ShutdownTimeout,
	}

	c.Logger.Debug(context.Background(), "HTTP server initialized", zap.String("address", c.Config.Server.GetServerAddr()))
}

// ---------------------- Command Handlers ----------------------

func (c *Container) createRegisterUserHandler() messaging.CommandHandler {
	handler := appHandlers.NewRegisterUserHandler(
		c.UserRepo,
		c.OutboxRepo,
		c.UnitOfWork,
		c.EventSerializer,
		c.Logger,
	)

	return messaging.CommandHandlerFunc(func(ctx context.Context, cmd messaging.Command) error {
		registerCmd, ok := cmd.(commands.RegisterUserCommand)
		if !ok {
			return fmt.Errorf("invalid command type: %T", cmd)
		}
		return handler.Handle(ctx, registerCmd)
	})
}

// ---------------------- HTTP Server Getter ----------------------

func (c *Container) HTTPServer() *http.Server {
	return c.httpServer
}

// Add more handler factories as you implement features:
/*

func (c *Container) createLoginHandler() messaging.CommandHandler {
	handler := appHandlers.NewLoginHandler(
		c.UserRepo,
		c.SessionRepo,
		c.TokenService,
		c.UnitOfWork,
		c.Logger,
	)

	return bus.CommandHandlerFunc(func(ctx context.Context, cmd messaging.Command) error {
		loginCmd := cmd.(commands.LoginCommand)
		return handler.Handle(ctx, loginCmd)
	})
}

func (c *Container) createVerifyEmailHandler() messaging.CommandHandler {
	handler := appHandlers.NewVerifyEmailHandler(
		c.UserRepo,
		c.EmailTokenRepo,
		c.UnitOfWork,
		c.Logger,
	)

	return bus.CommandHandlerFunc(func(ctx context.Context, cmd messaging.Command) error {
		verifyCmd := cmd.(commands.VerifyEmailCommand)
		return handler.Handle(ctx, verifyCmd)
	})
}

func (c *Container) createPasswordResetHandler() messaging.CommandHandler {
	handler := appHandlers.NewPasswordResetHandler(
		c.UserRepo,
		c.PasswordResetTokenRepo,
		c.EmailService,
		c.UnitOfWork,
		c.Logger,
	)

	return bus.CommandHandlerFunc(func(ctx context.Context, cmd messaging.Command) error {
		resetCmd := cmd.(commands.RequestPasswordResetCommand)
		return handler.Handle(ctx, resetCmd)
	})
}

func (c *Container) createChangePasswordHandler() messaging.CommandHandler {
	handler := appHandlers.NewChangePasswordHandler(
		c.UserRepo,
		c.UnitOfWork,
		c.Logger,
	)

	return bus.CommandHandlerFunc(func(ctx context.Context, cmd messaging.Command) error {
		changeCmd := cmd.(commands.ChangePasswordCommand)
		return handler.Handle(ctx, changeCmd)
	})
}
*/
