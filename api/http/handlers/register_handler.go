package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	apiDtos "authentication/api/http/dtos"
	"authentication/api/http/dtos/auth/request"
	"authentication/api/http/dtos/auth/response"
	"authentication/internal/application/commands"
	appDtos "authentication/internal/application/dtos"
	"authentication/internal/application/messaging"
	"authentication/internal/domain"
	"authentication/shared/logging"
	"authentication/shared/utils"

	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)


type AuthHandler struct {
	commandBus *messaging.CommandBus
	logger     logging.Logger
	validator  *validator.Validate
}

func NewAuthHandler(commandBus *messaging.CommandBus, logger logging.Logger) *AuthHandler {
	return &AuthHandler{
		commandBus: commandBus,
		logger:     logger.With(zap.String("handler", "auth")),
		validator:  utils.NewValidator(),
	}
}

func (h *AuthHandler) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.EmailRegistrationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(h.validator); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()
	cmd := req.ToCommand(clientIP, userAgent)

	appResult, err := messaging.Execute[commands.RegisterEmailUserCommand, appDtos.RegisterEmailUserResult](
		h.commandBus,
		ctx,
		cmd,
	)

	if err != nil {
		statusCode, message := h.mapErrorToHTTP(ctx, err)
		h.respondError(w, statusCode, message)
		return
	}

	httpResponse := response.EmailRegistrationResponse{
		UserID:          appResult.UserID,
		Email:           appResult.Email,
		Username:        appResult.Username,
		FirstName:       appResult.FirstName,
		LastName:        appResult.LastName,
		Role:            appResult.Role,
		RequiresOnboard: appResult.RequiresOnboard,
		CreatedAt:       time.Now().UTC(),
	}

	h.respondSuccess(w, http.StatusCreated, "Registration successful. Please check your email to verify your account.", httpResponse)
}

func (h *AuthHandler) RegisterOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.OAuthRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()
	cmd := req.ToCommand(clientIP, userAgent)

	// Execute through command bus - middleware handles logging, tracing, metrics
	appResult, err := messaging.Execute[commands.RegisterOAuthUserCommand, appDtos.RegisterOAuthUserResult](
		h.commandBus,
		ctx,
		cmd,
	)

	if err != nil {
		statusCode, message := h.mapErrorToHTTP(ctx, err)
		h.respondError(w, statusCode, message)
		return
	}

	var message string
	var statusCode int
	if appResult.IsNewUser {
		message = "Registration successful. Welcome!"
		statusCode = http.StatusCreated
	} else {
		message = "Login successful. Welcome back!"
		statusCode = http.StatusOK
	}

	httpResponse := response.OAuthRegistrationResponse{
		UserID:          appResult.UserID,
		Email:           appResult.Email,
		FirstName:       appResult.FirstName,
		LastName:        appResult.LastName,
		Role:            appResult.Role,
		IsNewUser:       appResult.IsNewUser,
		OAuthProvider:   appResult.OAuthProvider,
		RequiresOnboard: appResult.RequiresOnboard,
		CreatedAt:       time.Now().UTC(),
		AccessToken:     appResult.AccessToken,
		RefreshToken:    appResult.RefreshToken,
		TokenType:       "Bearer",
		ExpiresIn:       appResult.ExpiresIn,
		ExpiresAt:       appResult.ExpiresAt,
		SessionID:       appResult.SessionID,
	}

	h.respondSuccess(w, statusCode, message, httpResponse)
}

func (h *AuthHandler) respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(apiDtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func (h *AuthHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(apiDtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}

func (h *AuthHandler) mapErrorToHTTP(ctx context.Context, err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyInUse):
		return http.StatusConflict, "Email is already registered"
	case errors.Is(err, domain.ErrInvalidEmail):
		return http.StatusBadRequest, "Invalid email format"
	case errors.Is(err, domain.ErrInvalidUsernameFormat):
		return http.StatusBadRequest, "Invalid username format"
	case errors.Is(err, domain.ErrInvalidPassword):
		return http.StatusBadRequest, "Invalid password format"
	case errors.Is(err, domain.ErrInvalidRole):
		return http.StatusBadRequest, "Invalid role"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, "User not found"
	default:
		h.logger.Error(ctx, "Unexpected error", zap.Error(err))
		return http.StatusInternalServerError, "An unexpected error occurred"
	}
}