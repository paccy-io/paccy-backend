package main

import (
	"github.com/paccy-io/paccy-backend/pkg/logger"
)

func testWebhooks() {
	// Initialize logger with Discord webhook
	logger.Initialize("info", "development", "YOUR_DISCORD_WEBHOOK_URL_HERE")
	
	// Test startup notification
	logger.SendStartupNotification("Paccy Backend", "1.0.0", "development", "3000")
	
	// Test user registration notification
	logger.SendUserRegistrationNotification("6898738db444555c0ad490d4", "testuser", "test@example.com")
	
	// Test email verification notification
	logger.SendEmailVerificationNotification("6898738db444555c0ad490d4", "test@example.com")
	
	// Test email success notification
	logger.SendEmailSuccessNotification("6898738db444555c0ad490d4", "test@example.com", "Welcome to Paccy", "welcome_email", "resender")
	
	// Test email failure notification
	logger.SendEmailFailureNotification("6898738db444555c0ad490d4", "test@example.com", "Welcome to Paccy", "welcome_email", "resender", "SMTP connection timeout")
	
	// Regular server logger
	serverLogger := logger.Server()
	serverLogger.Info("All webhook notifications sent to Discord")
}