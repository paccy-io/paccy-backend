package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/auth"
	"github.com/paccy-io/paccy-backend/internal/user"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	responses "github.com/paccy-io/paccy-backend/pkg/response"
)

// JWTMiddleware verifies JWT token and sets user claims in context
func JWTMiddleware(authService auth.AuthService) fiber.Handler {
	log := logger.Auth()

	return func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id").(string)
		logCtx := log.WithRequestID(requestID)

		logCtx.Info("JWT Middleware called")

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logCtx.Warn("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Authorization header required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logCtx.Warn("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid authorization header format",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logCtx.Warn("Empty JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Empty token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify token
		claims, err := authService.VerifyToken(c.Context(), token)
		if err != nil {
			logCtx.WithError(err).Error("JWT token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Store claims in context for use in handlers
		c.Locals("user_claims", claims)
		c.Locals("user_id", claims.ID.Hex())

		logCtx.WithUserID(claims.ID.Hex()).Debug("JWT token verified successfully")
		return c.Next()
	}
}

// JWTWithUserValidationMiddleware verifies JWT token AND validates user exists in database
func JWTWithUserValidationMiddleware(authService auth.AuthService, userService user.UserService) fiber.Handler {
	log := logger.Auth()

	return func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id").(string)
		logCtx := log.WithRequestID(requestID)

		logCtx.Info("JWT with User Validation Middleware called")

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logCtx.Warn("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Authorization header required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logCtx.Warn("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid authorization header format",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logCtx.Warn("Empty JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Empty token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify token
		claims, err := authService.VerifyToken(c.Context(), token)
		if err != nil {
			logCtx.WithError(err).Error("JWT token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify user exists in database
		userID := claims.ID.Hex()
		_, err = userService.GetByID(c.Context(), userID)
		if err != nil {
			logCtx.WithField("user_id", userID).
				WithError(err).
				Warn("JWT token contains user that doesn't exist in database")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Store claims in context for use in handlers
		c.Locals("user_claims", claims)
		c.Locals("user_id", userID)

		logCtx.WithUserID(userID).Debug("JWT token and user existence verified successfully")
		return c.Next()
	}
}

// UserIDMiddleware verifies that the user ID in route matches the authenticated user
func UserIDMiddleware() fiber.Handler {
	log := logger.Auth()

	return func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id").(string)
		logCtx := log.WithRequestID(requestID)

		// Get user ID from route params
		routeUserID := c.Params("userID")
		logCtx.WithField("route_user_id", routeUserID).Debug("Extracting userID from route params")
		if routeUserID == "" {
			logCtx.Error("Missing userID in route parameters")
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User ID required in route",
				Code:    fiber.StatusBadRequest,
			})
		}

		// Get authenticated user ID from JWT claims (set by JWTMiddleware)
		authenticatedUserID := c.Locals("user_id")
		if authenticatedUserID == nil {
			logCtx.Error("No authenticated user found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Authentication required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Compare user IDs
		if routeUserID != authenticatedUserID.(string) {
			logCtx.WithField("route_user_id", routeUserID).
				WithField("authenticated_user_id", authenticatedUserID).
				Warn("User ID mismatch - unauthorized access attempt")
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Access denied: you can only access your own resources",
				Code:    fiber.StatusForbidden,
			})
		}

		logCtx.WithUserID(routeUserID).Debug("User ID authorization successful")
		return c.Next()
	}
}

// CombinedJWTUserIDMiddleware combines JWT verification and user ID validation
func CombinedJWTUserIDMiddleware(authService auth.AuthService) fiber.Handler {
	log := logger.Auth()

	return func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id").(string)
		logCtx := log.WithRequestID(requestID)

		logCtx.WithField("method", c.Method()).
			WithField("url", c.OriginalURL()).
			WithField("path", c.Path()).
			Info("Combined JWT + UserID middleware called")

		// First verify JWT
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logCtx.Warn("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Authorization header required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logCtx.Warn("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid authorization header format",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logCtx.Warn("Empty JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Empty token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify token
		claims, err := authService.VerifyToken(c.Context(), token)
		if err != nil {
			logCtx.WithError(err).Error("JWT token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Store claims in context for use in handlers
		c.Locals("user_claims", claims)
		c.Locals("user_id", claims.ID.Hex())

		logCtx.WithUserID(claims.ID.Hex()).Debug("JWT token verified successfully")

		// Debug: Show all route parameters
		logCtx.WithField("original_url", c.OriginalURL()).
			WithField("path", c.Path()).
			WithField("route", c.Route().Path).
			Info("Route debugging info")

		// Now verify user ID matches
		routeUserID := c.Params("userID")
		logCtx.WithField("route_user_id", routeUserID).Info("Extracting userID from route params")

		if routeUserID == "" {
			logCtx.Error("Missing userID in route parameters")
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User ID required in route",
				Code:    fiber.StatusBadRequest,
			})
		}

		// Compare user IDs
		if routeUserID != claims.ID.Hex() {
			logCtx.WithField("route_user_id", routeUserID).
				WithField("authenticated_user_id", claims.ID.Hex()).
				Warn("User ID mismatch - unauthorized access attempt")
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Access denied: you can only access your own resources",
				Code:    fiber.StatusForbidden,
			})
		}

		logCtx.WithUserID(routeUserID).Debug("Combined JWT + UserID authorization successful")
		return c.Next()
	}
}

// CombinedJWTUserIDWithValidationMiddleware combines JWT verification, user ID validation, and user existence check
func CombinedJWTUserIDWithValidationMiddleware(authService auth.AuthService, userService user.UserService) fiber.Handler {
	log := logger.Auth()

	return func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id").(string)
		logCtx := log.WithRequestID(requestID)

		logCtx.WithField("method", c.Method()).
			WithField("url", c.OriginalURL()).
			WithField("path", c.Path()).
			Info("Combined JWT + UserID + User Existence middleware called")

		// First verify JWT
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logCtx.Warn("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Authorization header required",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Check if it's Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logCtx.Warn("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid authorization header format",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logCtx.Warn("Empty JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Empty token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify token
		claims, err := authService.VerifyToken(c.Context(), token)
		if err != nil {
			logCtx.WithError(err).Error("JWT token verification failed")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Verify user exists in database
		userID := claims.ID.Hex()
		_, err = userService.GetByID(c.Context(), userID)
		if err != nil {
			logCtx.WithField("user_id", userID).
				WithError(err).
				Warn("JWT token contains user that doesn't exist in database")
			return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid or expired token",
				Code:    fiber.StatusUnauthorized,
			})
		}

		// Store claims in context for use in handlers
		c.Locals("user_claims", claims)
		c.Locals("user_id", userID)

		logCtx.WithUserID(userID).Debug("JWT token and user existence verified successfully")

		// Now verify user ID matches route parameter
		routeUserID := c.Params("userID")
		logCtx.WithField("route_user_id", routeUserID).Info("Extracting userID from route params")

		if routeUserID == "" {
			logCtx.Error("Missing userID in route parameters")
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User ID required in route",
				Code:    fiber.StatusBadRequest,
			})
		}

		// Compare user IDs
		if routeUserID != userID {
			logCtx.WithField("route_user_id", routeUserID).
				WithField("authenticated_user_id", userID).
				Warn("User ID mismatch - unauthorized access attempt")
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Access denied: you can only access your own resources",
				Code:    fiber.StatusForbidden,
			})
		}

		logCtx.WithUserID(routeUserID).Debug("Combined JWT + UserID + User Existence authorization successful")
		return c.Next()
	}
}
