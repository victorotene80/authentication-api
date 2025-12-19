package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"authentication/api/http/dtos"
	"authentication/api/http/dtos/auth/request"
	"authentication/api/http/dtos/auth/response"
	"authentication/internal/application/contracts/messaging"
	appDtos "authentication/internal/application/dtos"
	"authentication/internal/domain"
	"authentication/shared/logging"
	"authentication/shared/utils"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type LoginHandler struct {
	commandBus messaging.CommandBus
	logger     logging.Logger
	validator  *validator.Validate
}

func NewLoginHandler(commandBus messaging.CommandBus, logger logging.Logger) *LoginHandler {
	return &LoginHandler{
		commandBus: commandBus,
		logger:     logger.With(zap.String("handler", "login")),
		validator:  utils.NewValidator(),
	}
}

func (h *LoginHandler) LoginEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.EmailLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode email login request", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn(ctx, "Validation failed",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()
	cmd := req.ToCommand(clientIP, userAgent)

	h.logger.Info(ctx, "Processing email login",
		zap.String("email", req.Email),
	)

	// Create a context that can store the result
	resultCtx := ctx

	if err := h.commandBus.Execute(resultCtx, cmd); err != nil {
		h.logger.Error(ctx, "Email login failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)

		statusCode, message := h.mapErrorToHTTP(err)
		h.respondError(w, statusCode, message)
		return
	}

	// Retrieve result from context (set by handler)
	result, ok := resultCtx.Value("login_result").(*appDtos.LoginResult)
	if !ok {
		h.logger.Error(ctx, "Failed to retrieve login result from context")
		h.respondError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	resp := h.buildLoginResponse(result)

	h.logger.Info(ctx, "User logged in successfully via email",
		zap.String("user_id", result.UserID),
		zap.String("email", req.Email),
	)

	h.respondSuccess(w, http.StatusOK, "Login successful", resp)
}

func (h *LoginHandler) LoginOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.OAuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode OAuth login request", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn(ctx, "OAuth login validation failed",
			zap.Error(err),
		)
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()
	cmd := req.ToCommand(clientIP, userAgent)

	/*h.logger.Info(ctx, "Processing OAuth login",
		zap.String("provider", req.IDToken),
	)*/

	// Create a context that can store the result
	resultCtx := ctx

	if err := h.commandBus.Execute(resultCtx, cmd); err != nil {
		h.logger.Error(ctx, "OAuth login failed",
			//zap.String("provider", req.OAuthProvider),
			zap.Error(err),
		)

		statusCode, message := h.mapErrorToHTTP(err)
		h.respondError(w, statusCode, message)
		return
	}

	// Retrieve result from context (set by handler)
	result, ok := resultCtx.Value("login_result").(appDtos.LoginResult)
	if !ok {
		h.logger.Error(ctx, "Failed to retrieve login result from context")
		h.respondError(w, http.StatusInternalServerError, "Internal error")
		return
	}

	resp := h.buildOAuthLoginResponse(result)

	if result.IsNewUser {
		h.logger.Info(ctx, "New user registered and logged in via OAuth",
			zap.String("user_id", result.UserID),
			//zap.String("provider", req.OAuthProvider),
		)
	} else {
		h.logger.Info(ctx, "User logged in successfully via OAuth",
			zap.String("user_id", result.UserID),
			//zap.String("provider", req.OAuthProvider),
		)
	}

	message := "Login successful"
	if result.IsNewUser {
		message = "Registration and login successful"
	}

	h.respondSuccess(w, http.StatusOK, message, resp)
}

func (h *LoginHandler) buildLoginResponse(result *appDtos.LoginResult) response.LoginResponse {
	return response.LoginResponse{
		User: response.UserInfo{
			UserID:     result.UserID,
			Email:      result.Email,
			Username:   result.Username,
			FirstName:  result.FirstName,
			LastName:   result.LastName,
			Role:       result.Role,
			IsVerified: true, // Assuming verified if they can login
		},
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    result.ExpiresAt - utils.NowUTC().Unix(),
	}
}

func (h *LoginHandler) buildOAuthLoginResponse(result *appDtos.LoginResult) response.OAuthLoginResponse {
	return response.OAuthLoginResponse{
		LoginResponse: h.buildLoginResponse(result),
		IsNewUser:     result.IsNewUser,
	}
}

// mapErrorToHTTP converts domain errors to HTTP status codes and messages
func (h *LoginHandler) mapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Invalid email or password"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusUnauthorized, "Invalid email or password"
	case errors.Is(err, domain.ErrUserLocked):
		return http.StatusForbidden, "Account is locked"
	case errors.Is(err, domain.ErrOAuthVerificationFailed):
		return http.StatusUnauthorized, "OAuth verification failed"
	case errors.Is(err, domain.ErrOAuthEmailNotVerified):
		return http.StatusForbidden, "Email not verified by OAuth provider"
	case errors.Is(err, domain.ErrInvalidEmail):
		return http.StatusBadRequest, "Invalid email format"
	default:
		// Log unexpected errors
		h.logger.Error(context.Background(), "Unexpected error during login", zap.Error(err))
		return http.StatusInternalServerError, "An unexpected error occurred"
	}
}

func (h *LoginHandler) respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func (h *LoginHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}
