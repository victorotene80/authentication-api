package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens for protected routes
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Token is required",
				},
			})
			c.Abort()
			return
		}

		// TODO: Validate token using JWT service
		// For now, we'll just pass through
		// In a real implementation, you would:
		// 1. Parse and validate the JWT token
		// 2. Extract user information from the token
		// 3. Set user context in gin.Context
		// 
		// Example:
		// claims, err := jwtService.ValidateToken(token)
		// if err != nil {
		//     c.JSON(http.StatusUnauthorized, gin.H{...})
		//     c.Abort()
		//     return
		// }
		// c.Set("user_id", claims.UserID)
		// c.Set("email", claims.Email)

		c.Next()
	}
}

// OptionalAuthMiddleware tries to extract user info but doesn't require authentication
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			
			// TODO: Validate token and set user context if valid
			// If validation fails, just continue without user context
			_ = token
		}

		c.Next()
	}
}