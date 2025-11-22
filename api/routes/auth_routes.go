package routes

import (
	"authentication/api/handlers"
	"authentication/api/middleware"
	
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	authGroup := router.Group("/api/v1/auth")
	{
		// Public
		authGroup.POST("/register", authHandler.RegisterUser)
		/*authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/verify-email", authHandler.VerifyEmail)
		authGroup.POST("/resend-verification", authHandler.ResendVerification)
		authGroup.POST("/password-reset/request", authHandler.RequestPasswordReset)
		authGroup.POST("/password-reset/confirm", authHandler.ConfirmPasswordReset)*/
	}

	protectedGroup := router.Group("/api/v1/auth")
	protectedGroup.Use(middleware.AuthMiddleware())
	{
		/*protectedGroup.POST("/logout", authHandler.Logout)
		protectedGroup.POST("/password/change", authHandler.ChangePassword)
		protectedGroup.GET("/session", authHandler.GetSession)
		protectedGroup.DELETE("/session/:session_id", authHandler.RevokeSession)
		protectedGroup.DELETE("/sessions", authHandler.RevokeAllSessions)*/
	}
}
