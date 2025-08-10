package main

import (
	"github.com/paccy-io/paccy-backend/pkg/logger"
)

func main() {
	// Initialize logger with Discord webhook
	logger.Initialize("debug", "development", "YOUR_DISCORD_WEBHOOK_URL_HERE")

	// Test different log levels
	authLogger := logger.Auth()
	userLogger := logger.User()
	notifyLogger := logger.Notify()

	// These will show up normally
	authLogger.Info("User login successful")
	userLogger.WithField("userID", "123").Debug("User profile retrieved")
	notifyLogger.WithField("email", "test@example.com").Info("Email sent successfully")

	// These will trigger Discord webhook (if configured)
	authLogger.WithField("userID", "456").WithField("ip", "192.168.1.1").Error("Failed login attempt detected")
	userLogger.WithField("userID", "789").Error("Database connection failed")
}
