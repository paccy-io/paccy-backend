package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/paccy-io/paccy-backend/internal/user"

	//"github.com/paccy-io/paccy-backend/internal/user/dto"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
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
	id := c.Params("id")

	var _user *models.SafeUser
	var err error

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if _, objErr := primitive.ObjectIDFromHex(id); objErr == nil {
		// Get user by MongoDB ObjectID
		_user, err = h.userService.GetByID(ctx, id)
	} else if _, uuidErr := uuid.Parse(id); uuidErr == nil {
		// Get user by UUID
		_user, err = h.userService.GetByUUID(ctx, id)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID format",
			Code:    fiber.StatusBadRequest,
		})
	}

	if err != nil {
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

	response := responses.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    _user,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// jwt support implementation before moving on to the next handler (// UpdateUser, DeleteUser, etc.)
