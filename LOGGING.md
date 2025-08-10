# Modern Logging System Usage

This project now includes a production-level structured logging system with service-specific prefixes.

## Features

- **Service-specific prefixes**: `[auth]`, `[user]`, `[notify]`, `[mongo]`, etc.
- **Structured logging**: JSON format in production, human-readable in development
- **Request tracking**: Automatic request ID generation and logging
- **Context support**: Add user IDs, request IDs, and other context to logs
- **Multiple log levels**: Debug, Info, Warn, Error, Fatal

## Usage Examples

### Basic Service Logging

```go
// In auth service
authLogger := logger.Auth()
authLogger.Info("User login attempt")
authLogger.WithField("email", "user@example.com").Info("Login successful")
authLogger.WithError(err).Error("Login failed")

// In user service
userLogger := logger.User()
userLogger.WithUserID("12345").Info("User profile updated")

// In notify service
notifyLogger := logger.Notify()
notifyLogger.WithField("provider", "resender").Info("Email sent successfully")

// In mongo operations
mongoLogger := logger.Mongo()
mongoLogger.WithField("collection", "users").Debug("Executing query")
```

### Context Logging

```go
// Add multiple context fields
logger := logger.Auth().WithContext(map[string]interface{}{
    "user_id": "12345",
    "action": "password_reset",
    "ip": "192.168.1.1",
})
logger.Info("Password reset initiated")

// Chain context
logger.Auth().
    WithUserID("12345").
    WithRequestID("req-67890").
    WithField("provider", "google").
    Info("OAuth login successful")
```

### HTTP Request Logging

The middleware automatically logs all HTTP requests with:

- Method, path, status code
- Request duration
- IP address, User-Agent
- Request and response sizes
- Request ID for tracing

### Error Logging

```go
if err != nil {
    logger.Auth().
        WithError(err).
        WithField("user_id", userID).
        Error("Failed to authenticate user")
}
```

## Log Output Examples

### Development (Human-readable)

```
time="2025-01-01 12:00:00" level=info msg="User login attempt" service=auth email=user@example.com
time="2025-01-01 12:00:01" level=info msg="Request completed" service=http method=POST path=/login status=200 duration=150 ip=192.168.1.1
```

### Production (JSON)

```json
{
  "timestamp": "2025-01-01T12:00:00.000Z",
  "level": "info",
  "message": "User login attempt",
  "service": "auth",
  "email": "user@example.com"
}
```

## Configuration

The logging system is automatically initialized in `config.LoadConfig()` using:

- `APP_LOGGING_LEVEL`: debug, info, warn, error, fatal
- `APP_ENVIRONMENT`: development (text format) or production (JSON format)

## Available Loggers

- `logger.Auth()` - Authentication operations
- `logger.User()` - User management
- `logger.Notify()` - Email/notification services
- `logger.Mongo()` - Database operations
- `logger.Server()` - Server-level operations
- `logger.Config()` - Configuration loading
- `logger.HTTP()` - HTTP requests (used by middleware)
- `logger.JWT()` - JWT operations
