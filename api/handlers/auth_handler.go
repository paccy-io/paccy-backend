package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/auth"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	responses "github.com/paccy-io/paccy-backend/pkg/response"

	authDto "github.com/paccy-io/paccy-backend/internal/auth/dto"
	"github.com/paccy-io/paccy-backend/internal/user/dto"
)

type AuthHandler struct {
	authService auth.AuthService
	log         *logger.Logger
}

func NewAuthHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         logger.Auth(),
	}
}

func (h *AuthHandler) RegisterAndSendVerification(c *fiber.Ctx) error {
	requestID := c.Locals("request_id").(string)
	logCtx := h.log.WithRequestID(requestID)

	dtoUser := new(dto.CreateUserDTO)
	if err := c.BodyParser(dtoUser); err != nil {
		logCtx.WithError(err).Error("Failed to parse request body for user registration")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request body: %v", err),
			Code:    fiber.StatusBadRequest,
		})
	}

	logCtx.WithField("email", dtoUser.Email).Info("Attempting user registration")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	bool, err := h.authService.RegisterAndSendVerification(ctx, dtoUser)
	if err != nil {
		if bool {
			logCtx.WithError(err).WithField("email", dtoUser.Email).Warn("User registered but email verification failed")
			return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
				Success: true,
				Error:   fmt.Sprintf("Failed to send email verification: %v", err),
				Message: "User registered but sending email verification failed",
			})
		} else {
			logCtx.WithError(err).WithField("email", dtoUser.Email).Error("User registration failed")
			return c.Status(fiber.ErrBadRequest.Code).JSON(responses.ErrorResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to register: %v", err),
				Code:    fiber.StatusInternalServerError,
			})
		}
	}

	logCtx.WithField("email", dtoUser.Email).Info("User registered and verification email sent successfully")
	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    nil,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginDTO authDto.LoginDTO
	if err := c.BodyParser(&loginDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request body: %v", err),
			Code:    fiber.StatusBadRequest,
		})
	}

	// validate the input
	email := loginDTO.Email
	password := loginDTO.Password

	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Email and password are required",
			Code:    fiber.StatusBadRequest,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	token, err := h.authService.Login(ctx, email, password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Login failed: %v", err),
			Code:    fiber.StatusUnauthorized,
		})
	}

	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    map[string]string{"token": token},
	})
}

func (h *AuthHandler) SendEmailVerification(c *fiber.Ctx) error {
	userID := c.Params("userID")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "User ID is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.authService.SendEmailVerification(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to send email verification: %v", err),
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Email verification sent successfully",
		Data:    nil,
	})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	requestID := c.Locals("request_id").(string)
	logCtx := h.log.WithRequestID(requestID)

	token := c.Params("token")
	if token == "" {
		logCtx.Warn("Email verification attempted without token")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Token is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	logCtx.WithField("token_length", len(token)).Info("Attempting email verification")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.authService.VerifyEmail(ctx, token)
	if err != nil {
		logCtx.WithError(err).Error("Email verification failed")
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to verify email: %v", err),
			Code:    fiber.StatusInternalServerError,
		})
	}

	logCtx.Info("Email verified successfully")
	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Email verified successfully",
		Data:    nil,
	})
}

func (h *AuthHandler) VerifyToken(c *fiber.Ctx) error {
	// bearer token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Missing or invalid Authorization header",
		})
	}

	token := authHeader[len("Bearer "):]
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	claims, err := h.authService.VerifyToken(ctx, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Token verification failed: %v", err),
			Code:    fiber.StatusUnauthorized,
		})
	}

	return c.Status(fiber.StatusOK).JSON(responses.APIResponse{
		Success: true,
		Message: "Token verified successfully",
		Data:    claims,
	})
}
