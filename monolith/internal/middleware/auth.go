package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kanchan755/wallet_app/monolith/internal/auth"
	customError "github.com/kanchan755/wallet_app/monolith/internal/errors"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "NO TOKEN PROVIDE", "no token provided"))
			c.Abort()
			return
		}

		//split Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "Invalid token header", "invalid header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Error(customError.NewAppError(http.StatusUnauthorized, "Invalid token", "invalid token or expired"))
			c.Abort()
			return
		}

		//attach user to context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

	}
}
