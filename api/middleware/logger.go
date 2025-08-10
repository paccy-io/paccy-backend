package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/pkg/logger"
)

// LoggerMiddleware creates a Fiber middleware for structured logging
func LoggerMiddleware() fiber.Handler {
	httpLogger := logger.HTTP()

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Log request details
		logEntry := httpLogger.WithContext(map[string]interface{}{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"duration":   duration.Milliseconds(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
			"bytes_in":   len(c.Body()),
			"bytes_out":  len(c.Response().Body()),
		})

		// Add request ID if available
		if requestID := c.Get("X-Request-ID"); requestID != "" {
			logEntry = logEntry.WithRequestID(requestID)
		}

		// Log based on status code
		status := c.Response().StatusCode()
		switch {
		case status >= 500:
			logEntry.Error("Internal server error")
		case status >= 400:
			logEntry.Warn("Client error")
		case status >= 300:
			logEntry.Info("Redirection")
		default:
			logEntry.Info("Request completed")
		}

		return err
	}
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (you can use UUID library for better IDs)
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Store in context for use in handlers
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// Simple request ID generator (replace with UUID for production)
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + time.Now().Format("000000")
}
