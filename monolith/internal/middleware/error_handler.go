package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customErrors "github.com/kanchan755/wallet_app/monolith/internal/errors"
	"github.com/kanchan755/wallet_app/monolith/internal/logger"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process the request

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if appErr, ok := err.(*customErrors.AppError); ok {
			logger.Warn(c.Request.Context(), "application error", "code", appErr.Code, "message", appErr.Message, "status_code", appErr.StatusCode)
			c.JSON(appErr.StatusCode, gin.H{
				"code":        appErr.Code,
				"message":     appErr.Message,
				"status_code": appErr.StatusCode,
			})
			return
		}

		// if error is not cover from our custom error, we will log it and return a generic internal server error response

		logger.Error(c.Request.Context(), "unexpected error", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
