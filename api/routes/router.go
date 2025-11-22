package routes

import (
	"authentication/api/handlers"
	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Setup auth routes
	SetupAuthRoutes(router, authHandler)

	return router
}
