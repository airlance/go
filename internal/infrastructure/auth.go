package infrastructure

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func AuthMiddleware(validTokens []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		if !isValidToken(token, validTokens) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isValidToken(token string, validTokens []string) bool {
	for _, validToken := range validTokens {
		if token == validToken {
			return true
		}
	}
	return false
}
