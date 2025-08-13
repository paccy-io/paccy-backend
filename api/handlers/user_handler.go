package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/paccy-io/paccy-backend/internal/user"
	"github.com/paccy-io/paccy-backend/internal/user/dto"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	model "github.com/paccy-io/paccy-backend/pkg/models"
	responses "github.com/paccy-io/paccy-backend/pkg/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService user.UserService
}

func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	pagination := &model.PaginationQuery{
		Page:  page,
		Limit: limit,
	}

	users, total, err := h.userService.GetUsers(ctx, pagination)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve users",
			Code:    fiber.StatusInternalServerError,
		})
	}

	totalPages := (total + limit - 1) / limit

	response := responses.APIResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
		Meta: &responses.MetaResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("userID")

	// Add debug logging
	log := logger.User()
	log.WithField("user_id", id).Info("GetUser handler called")

	var _user *models.SafeUser
	var err error
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if _, objErr := primitive.ObjectIDFromHex(id); objErr == nil {
		// Get user by MongoDB ObjectID
		log.Info("Getting user by ObjectID")
		_user, err = h.userService.GetByID(ctx, id)
	} else if _, uuidErr := uuid.Parse(id); uuidErr == nil {
		// Get user by UUID
		log.Info("Getting user by UUID")
		_user, err = h.userService.GetByUUID(ctx, id)
	} else {
		log.Error("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
			Code:    fiber.StatusBadRequest,
		})
	}

	if err != nil {
		log.WithError(err).Error("Error retrieving user")
		if err == ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	log.WithField("user_found", _user != nil).Info("User retrieved successfully, preparing response")

	response := responses.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    _user,
	}

	log.Info("About to return response")
	return c.Status(fiber.StatusOK).JSON(response)
}

// UpdateUserFields updates basic user information (username, first name, last name)
func (h *UserHandler) UpdateUserFields(c *fiber.Ctx) error {
	log := logger.User()
	requestID := c.Locals("request_id").(string)
	logCtx := log.WithRequestID(requestID)

	userID := c.Params("userID")
	logCtx.WithUserID(userID).Info("Updating user fields")

	var updateDTO dto.UpdateUserDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		logCtx.WithError(err).Error("Failed to parse update user request body")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.userService.UpdateUserFields(ctx, userID, &updateDTO)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update user fields")

		switch err {
		case ErrUserNotFound:
			return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User not found",
				Code:    fiber.StatusNotFound,
			})
		case ErrUserExists:
			return c.Status(fiber.StatusConflict).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Username already exists",
				Code:    fiber.StatusConflict,
			})
		case ErrInvalidInput:
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
				Code:    fiber.StatusBadRequest,
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Failed to update user fields",
				Code:    fiber.StatusInternalServerError,
			})
		}
	}

	logCtx.WithUserID(userID).Info("User fields updated successfully")

	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "User fields updated successfully",
	})
}

// UpdateProfile updates user profile information (bio, avatar, graduation year, etc.)
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	log := logger.User()
	requestID := c.Locals("request_id").(string)
	logCtx := log.WithRequestID(requestID)

	userID := c.Params("userID")
	logCtx.WithUserID(userID).Info("Updating user profile")

	var updateDTO dto.UpdateUserProfileDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		logCtx.WithError(err).Error("Failed to parse update profile request body")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.userService.UpdateProfile(ctx, userID, &updateDTO)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update user profile")

		switch err {
		case ErrUserNotFound:
			return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User not found",
				Code:    fiber.StatusNotFound,
			})
		case ErrInvalidInput:
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
				Code:    fiber.StatusBadRequest,
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Failed to update user profile",
				Code:    fiber.StatusInternalServerError,
			})
		}
	}

	logCtx.WithUserID(userID).Info("User profile updated successfully")

	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "User profile updated successfully",
	})
}

// UpdatePreferences updates user preferences (theme, language, notifications, etc.)
func (h *UserHandler) UpdatePreferences(c *fiber.Ctx) error {
	log := logger.User()
	requestID := c.Locals("request_id").(string)
	logCtx := log.WithRequestID(requestID)

	userID := c.Params("userID")
	logCtx.WithUserID(userID).Info("Updating user preferences")

	var updateDTO dto.UpdateUserPreferencesDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		logCtx.WithError(err).Error("Failed to parse update preferences request body")
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.userService.UpdatePreferences(ctx, userID, &updateDTO)
	if err != nil {
		logCtx.WithError(err).Error("Failed to update user preferences")

		switch err {
		case ErrUserNotFound:
			return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "User not found",
				Code:    fiber.StatusNotFound,
			})
		case ErrInvalidInput:
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
				Code:    fiber.StatusBadRequest,
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
				Success: false,
				Error:   "Failed to update user preferences",
				Code:    fiber.StatusInternalServerError,
			})
		}
	}

	logCtx.WithUserID(userID).Info("User preferences updated successfully")

	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "User preferences updated successfully",
	})
}
