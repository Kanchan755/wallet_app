package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// read from request header
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			// If not present, generate a new correlation ID (you can implement your own logic)
			correlationID = uuid.New().String()

		}
		// Set the correlation ID in the request context
		ctx := context.WithValue(c.Request.Context(), logger.CorrelationIDKey, correlationID)
		c.Request = c.Request.WithContext(ctx)

		//set in response header
		c.Header("X-Correlation-ID", correlationID)

		// Call the next handler in the chain
		c.Next()

	}
}

// generateCorrelationID generates a unique correlation ID (you can implement your own logic)
