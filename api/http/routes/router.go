package routes

import (
	"net/http"

	"authentication/api/http/handlers"
	"authentication/internal/application/contracts/messaging"
	"authentication/shared/logging"

	"github.com/gorilla/mux"
)

func SetupAuthRoutes(
	router *mux.Router,
	commandBus messaging.CommandBus,
	logger logging.Logger,
) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(commandBus, logger)
	loginHandler := handlers.NewAuthHandler(commandBus, logger)

	// Auth subrouter
	authRouter := router.PathPrefix("/api/v1/auth").Subrouter()

	// Registration endpoints
	authRouter.HandleFunc("/register/email", authHandler.RegisterEmail).Methods(http.MethodPost)
	authRouter.HandleFunc("/register/oauth", authHandler.RegisterOAuth).Methods(http.MethodPost)

	// Login endpoints
	authRouter.HandleFunc("/login/email", loginHandler.LoginEmail).Methods(http.MethodPost)
	authRouter.HandleFunc("/login/oauth", loginHandler.LoginOAuth).Methods(http.MethodPost)

	// Additional endpoints (implement these later)
	// authRouter.HandleFunc("/logout", authHandler.Logout).Methods(http.MethodPost)
	// authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods(http.MethodPost)
	// authRouter.HandleFunc("/verify-email", authHandler.VerifyEmail).Methods(http.MethodPost)
	// authRouter.HandleFunc("/forgot-password", authHandler.ForgotPassword).Methods(http.MethodPost)
	// authRouter.HandleFunc("/reset-password", authHandler.ResetPassword).Methods(http.MethodPost)
}