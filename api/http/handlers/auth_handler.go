package handlers

import (
	"encoding/json"
	"net/http"

	"authentication/api/http/dtos"
	"authentication/api/http/dtos/auth/request"
	"authentication/api/http/dtos/auth/response"

	//"authentication/internal/application/commands"
	"authentication/internal/application/contracts/messaging"
	"authentication/shared/logging"
	"authentication/shared/utils"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AuthHandler struct {
	commandBus messaging.CommandBus
	logger     logging.Logger
	validator  *validator.Validate
}

func NewAuthHandler(commandBus messaging.CommandBus, logger logging.Logger) *AuthHandler {
	return &AuthHandler{
		commandBus: commandBus,
		logger:     logger.With(zap.String("handler", "auth")),
		validator:  utils.NewValidator(),
	}
}

func (h *AuthHandler) RegisterEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.EmailRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode email registeration request", zap.Error(err))
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

	h.logger.Info(ctx, "Processing email registration",
		zap.String("email", req.Email),
		zap.String("username", req.Username),
	)

	if err := h.commandBus.Execute(ctx, cmd); err != nil {
		h.logger.Error(ctx, "Email registration failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)

		return
	}

	/*if err := h.commandBus.Execute(ctx, cmd); err != nil {
		h.logger.Error(ctx, "Email registration failed", zap.Error(err))
		statusCode, message := h.mapErrorToHTTP(err)
		h.respondError(w, statusCode, message)
		return
	}*/

	resp := response.RegisterUserResponse{
		//Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user",
		CreatedAt: utils.NowUTC(),
		//IsOAuthUser:     false,
		//RequiresOnboard: false,
	}

	h.logger.Info(ctx, "User registered successfully via email",
		zap.String("email", req.Email),
		zap.String("username", req.Username),
	)

	h.respondSuccess(w, http.StatusCreated, "Registration successful. Please verify your email.", resp)
}

func (h *AuthHandler) RegisterOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req request.OAuthRegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "Failed to decode OAuth registration request", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.validator); err != nil {
		h.logger.Warn(ctx, "OAuth registration validation failed",
			zap.Error(err),
			zap.String("IDToken", req.IDToken),
		)
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	clientIP := utils.GetClientIP(r)
	userAgent := r.UserAgent()

	cmd := req.ToCommand(clientIP, userAgent)

	h.logger.Info(ctx, "Processing OAuth registration",
		zap.String("IDToken", req.IDToken),
	)

	if err := h.commandBus.Execute(ctx, cmd); err != nil {
		h.logger.Error(ctx, "OAuth registration failed",
			zap.String("IDToken", req.IDToken),
			zap.Error(err),
		)

		//statusCode, message := h.mapErrorToHTTP(err)
		//h.respondError(w, statusCode, message)
		return
	}

	//come back to this later
	resp := response.RegisterUserResponse{}

	h.respondSuccess(w, http.StatusCreated, "Registration and login successful", resp)

}

func (h *AuthHandler) respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func (h *AuthHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dtos.ApiResponse[interface{}]{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}
