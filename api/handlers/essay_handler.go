package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/essay"
	"github.com/paccy-io/paccy-backend/internal/essay/dto"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	response "github.com/paccy-io/paccy-backend/pkg/response"
)

var essayHandlerLogger = logger.New("essay-handler")

type EssayHandler struct {
	essayService essay.EssayService
}

func NewEssayHandler(essayService essay.EssayService) *EssayHandler {
	return &EssayHandler{
		essayService: essayService,
	}
}

// CreateEssay creates a new essay for the authenticated user
func (h *EssayHandler) CreateEssay(c *fiber.Ctx) error {
	// Get user ID from JWT claims (set by auth middleware)
	userID := c.Locals("user_id").(string)

	var createDTO dto.CreateEssayDTO
	if err := c.BodyParser(&createDTO); err != nil {
		essayHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essay, err := h.essayService.Create(ctx, userID, &createDTO)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		essayHandlerLogger.Error("Failed to create essay", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create essay",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Data:    essay,
		Message: "Essay created successfully",
	})
}

// GetEssay retrieves a specific essay by ID
func (h *EssayHandler) GetEssay(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	essayID := c.Params("id")

	if essayID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Essay ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essay, err := h.essayService.GetByID(ctx, userID, essayID)
	if err != nil {
		if err == errors.ErrGetFailed {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Essay not found",
			})
		}
		essayHandlerLogger.Error("Failed to get essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": essayID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essay",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essay,
	})
}

// GetUserEssays retrieves all essays for the authenticated user
func (h *EssayHandler) GetUserEssays(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essays, err := h.essayService.GetByUserID(ctx, userID)
	if err != nil {
		essayHandlerLogger.Error("Failed to get user essays", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essays",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essays,
		Message: "Essays retrieved successfully",
	})
}

// GetEssaysByCollegeApplication retrieves all essays for a specific college application
func (h *EssayHandler) GetEssaysByCollegeApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collegeApplicationID := c.Params("college_application_id")

	if collegeApplicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "College application ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essays, err := h.essayService.GetByCollegeApplicationID(ctx, userID, collegeApplicationID)
	if err != nil {
		if err == errors.ErrGetFailed {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "College application not found or access denied",
			})
		}
		essayHandlerLogger.Error("Failed to get essays by college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essays",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essays,
		Message: "Essays retrieved successfully",
	})
}

// UpdateEssay updates an existing essay
func (h *EssayHandler) UpdateEssay(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	essayID := c.Params("id")

	if essayID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Essay ID is required",
		})
	}

	var updateDTO dto.UpdateEssayDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		essayHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id":  userID,
			"essay_id": essayID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essay, err := h.essayService.Update(ctx, userID, essayID, &updateDTO)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		if err == errors.ErrGetFailed {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Essay not found",
			})
		}
		essayHandlerLogger.Error("Failed to update essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": essayID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update essay",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essay,
		Message: "Essay updated successfully",
	})
}

// DeleteEssay deletes an essay
func (h *EssayHandler) DeleteEssay(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	essayID := c.Params("id")

	if essayID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Essay ID is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.essayService.Delete(ctx, userID, essayID)
	if err != nil {
		if err == errors.ErrGetFailed {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Essay not found",
			})
		}
		essayHandlerLogger.Error("Failed to delete essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": essayID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete essay",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Message: "Essay deleted successfully",
	})
}

// GetEssaysByStatus retrieves essays by status for the authenticated user
func (h *EssayHandler) GetEssaysByStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	status := c.Query("status")

	if status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Status parameter is required",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essays, err := h.essayService.GetByStatus(ctx, userID, status)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid status value",
			})
		}
		essayHandlerLogger.Error("Failed to get essays by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essays",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essays,
		Message: "Essays retrieved successfully",
	})
}

// GetEssaysWithDeadlines retrieves essays with upcoming deadlines
func (h *EssayHandler) GetEssaysWithDeadlines(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	daysStr := c.Query("days", "30") // Default to 30 days

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid days parameter",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	essays, err := h.essayService.GetEssaysWithDeadlines(ctx, userID, days)
	if err != nil {
		essayHandlerLogger.Error("Failed to get essays with deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essays",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    essays,
		Message: "Essays with upcoming deadlines retrieved successfully",
	})
}

// GetEssayCount retrieves the count of essays for the authenticated user
func (h *EssayHandler) GetEssayCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	count, err := h.essayService.Count(ctx, userID)
	if err != nil {
		essayHandlerLogger.Error("Failed to get essay count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve essay count",
		})
	}

	return c.JSON(response.SuccessResponse{
		Success: true,
		Data:    map[string]int64{"count": count},
		Message: "Essay count retrieved successfully",
	})
}
