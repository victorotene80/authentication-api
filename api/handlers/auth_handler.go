package handlers

import (
	"authentication/api/dto"
	"authentication/internal/application/commands"
	appValidators "authentication/api/validators"
	bus "authentication/internal/application/contracts/messaging"
	"authentication/internal/domain"
	"authentication/pkg/httputil"
	"authentication/shared/logging"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	commandBus bus.CommandBus
	validator  *appValidators.AuthValidator
}

func NewAuthHandler(commandBus bus.CommandBus) *AuthHandler {
	return &AuthHandler{
		commandBus: commandBus,
		validator:  appValidators.NewAuthValidator(),
	}
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var req dto.RegisterUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logging.ErrorCtx(c.Request.Context(), "failed to bind request", zap.Error(err))
		httputil.RespondWithError(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	// Map DTO -> Command
	cmd := commands.RegisterUserCommand{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Early business-level validation (application validator) — quick fail before sending to bus
	if err := h.validator.ValidateRegistration(cmd); err != nil {
		logging.WarnCtx(c.Request.Context(), "validation failed", zap.Error(err))
		httputil.RespondWithError(c, http.StatusBadRequest, "validation failed", err)
		return
	}

	// Execute command via bus (which should route to application handler)
	if err := h.commandBus.Execute(c.Request.Context(), cmd); err != nil {
		logging.ErrorCtx(c.Request.Context(), "failed to register user",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.String("username", req.Username),
		)

		// Domain-specific errors -> friendly statuses
		if errors.Is(err, domain.ErrEmailAlreadyInUse) {
			httputil.RespondWithError(c, http.StatusConflict, "email already registered", err)
			return
		}
		if errors.Is(err, domain.ErrUsernameAlreadyTaken) {
			httputil.RespondWithError(c, http.StatusConflict, "username already taken", err)
			return
		}
		// Other domain errors can be handled similarly

		httputil.RespondWithError(c, http.StatusInternalServerError, "failed to register user", err)
		return
	}

	resp := dto.RegisterUserResponse{
		Message: "User registered successfully. Please check your email for verification.",
	}
	logging.InfoCtx(c.Request.Context(), "user registered successfully",
		zap.String("email", req.Email),
		zap.String("username", req.Username),
	)

	httputil.RespondWithSuccess(c, http.StatusCreated, resp.Message, nil)
}
