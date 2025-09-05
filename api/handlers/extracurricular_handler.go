package handlers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/extracurriculars"
	"github.com/paccy-io/paccy-backend/internal/extracurriculars/dto"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	response "github.com/paccy-io/paccy-backend/pkg/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var extracurricularHandlerLogger = logger.New("extracurricular_handler")

type ExtracurricularHandler struct {
	service   extracurriculars.ExtracurricularService
	validator *validator.Validate
}

func NewExtracurricularHandler(service extracurriculars.ExtracurricularService) *ExtracurricularHandler {
	return &ExtracurricularHandler{
		service:   service,
		validator: validator.New(),
	}
}

// CreateExtracurricular creates a new extracurricular activity
// @Summary Create extracurricular
// @Description Create a new extracurricular activity for the authenticated user
// @Tags extracurriculars
// @Accept json
// @Produce json
// @Param extracurricular body dto.CreateExtracurricularDTO true "Extracurricular data"
// @Success 201 {object} response.SuccessResponse{data=dto.ExtracurricularResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars [post]
func (h *ExtracurricularHandler) CreateExtracurricular(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		extracurricularHandlerLogger.Error("Invalid user ID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	var req dto.CreateExtracurricularDTO
	if err := c.BodyParser(&req); err != nil {
		extracurricularHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		extracurricularHandlerLogger.Error("Validation failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	extracurricular, err := h.service.CreateExtracurricular(ctx, userObjectID, &req)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to create extracurricular", map[string]interface{}{
			"error":         err.Error(),
			"user_id":       userID,
			"activity_name": req.ActivityName,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create extracurricular",
		})
	}

	extracurricularHandlerLogger.Info("Extracurricular created successfully", map[string]interface{}{
		"extracurricular_id": extracurricular.ID,
		"user_id":            userID,
		"activity_name":      extracurricular.ActivityName,
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurricular created successfully",
		Data:    extracurricular,
	})
}

// GetExtracurriculars retrieves all extracurriculars for the authenticated user
// @Summary Get user's extracurriculars
// @Description Get all extracurricular activities for the authenticated user
// @Tags extracurriculars
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]dto.ExtracurricularResponseDTO}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars [get]
func (h *ExtracurricularHandler) GetExtracurriculars(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)

	extracurriculars, err := h.service.GetExtracurricularsByUserID(ctx, userID)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to get extracurriculars", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get extracurriculars",
		})
	}

	extracurricularHandlerLogger.Info("Retrieved extracurriculars for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(extracurriculars),
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurriculars retrieved successfully",
		Data:    extracurriculars,
	})
}

// GetExtracurricularByID retrieves a specific extracurricular by ID
// @Summary Get extracurricular by ID
// @Description Get a specific extracurricular activity by its ID
// @Tags extracurriculars
// @Produce json
// @Param id path string true "Extracurricular ID"
// @Success 200 {object} response.SuccessResponse{data=dto.ExtracurricularResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars/{id} [get]
func (h *ExtracurricularHandler) GetExtracurricularByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)
	extracurricularID := c.Params("id")

	if extracurricularID == "" {
		extracurricularHandlerLogger.Error("Missing extracurricular ID", map[string]interface{}{
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Extracurricular ID is required",
		})
	}

	extracurricular, err := h.service.GetExtracurricularByID(ctx, extracurricularID, userID)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to get extracurricular by ID", map[string]interface{}{
			"error":              err.Error(),
			"user_id":            userID,
			"extracurricular_id": extracurricularID,
		})
		return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Extracurricular not found",
		})
	}

	extracurricularHandlerLogger.Info("Retrieved extracurricular by ID", map[string]interface{}{
		"user_id":            userID,
		"extracurricular_id": extracurricularID,
		"activity_name":      extracurricular.ActivityName,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurricular retrieved successfully",
		Data:    extracurricular,
	})
}

// GetExtracurricularsByCategory retrieves extracurriculars by category
// @Summary Get extracurriculars by category
// @Description Get extracurricular activities by category for the authenticated user
// @Tags extracurriculars
// @Produce json
// @Param category path string true "Extracurricular category"
// @Success 200 {object} response.SuccessResponse{data=[]dto.ExtracurricularResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars/category/{category} [get]
func (h *ExtracurricularHandler) GetExtracurricularsByCategory(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)
	category := c.Params("category")

	if category == "" {
		extracurricularHandlerLogger.Error("Missing category", map[string]interface{}{
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Category is required",
		})
	}

	extracurriculars, err := h.service.GetExtracurricularsByCategory(ctx, userID, category)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to get extracurriculars by category", map[string]interface{}{
			"error":    err.Error(),
			"user_id":  userID,
			"category": category,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get extracurriculars by category",
		})
	}

	extracurricularHandlerLogger.Info("Retrieved extracurriculars by category", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(extracurriculars),
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurriculars retrieved successfully",
		Data:    extracurriculars,
	})
}

// UpdateExtracurricular updates an existing extracurricular
// @Summary Update extracurricular
// @Description Update an existing extracurricular activity
// @Tags extracurriculars
// @Accept json
// @Produce json
// @Param id path string true "Extracurricular ID"
// @Param extracurricular body dto.UpdateExtracurricularDTO true "Extracurricular update data"
// @Success 200 {object} response.SuccessResponse{data=dto.ExtracurricularResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars/{id} [put]
func (h *ExtracurricularHandler) UpdateExtracurricular(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)
	extracurricularID := c.Params("id")

	if extracurricularID == "" {
		extracurricularHandlerLogger.Error("Missing extracurricular ID for update", map[string]interface{}{
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Extracurricular ID is required",
		})
	}

	var req dto.UpdateExtracurricularDTO
	if err := c.BodyParser(&req); err != nil {
		extracurricularHandlerLogger.Error("Failed to parse update request body", map[string]interface{}{
			"error":              err.Error(),
			"user_id":            userID,
			"extracurricular_id": extracurricularID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		extracurricularHandlerLogger.Error("Update validation failed", map[string]interface{}{
			"error":              err.Error(),
			"user_id":            userID,
			"extracurricular_id": extracurricularID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	extracurricular, err := h.service.UpdateExtracurricular(ctx, extracurricularID, userID, &req)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to update extracurricular", map[string]interface{}{
			"error":              err.Error(),
			"user_id":            userID,
			"extracurricular_id": extracurricularID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update extracurricular",
		})
	}

	extracurricularHandlerLogger.Info("Extracurricular updated successfully", map[string]interface{}{
		"extracurricular_id": extracurricularID,
		"user_id":            userID,
		"activity_name":      extracurricular.ActivityName,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurricular updated successfully",
		Data:    extracurricular,
	})
}

// DeleteExtracurricular deletes an extracurricular
// @Summary Delete extracurricular
// @Description Delete an extracurricular activity
// @Tags extracurriculars
// @Produce json
// @Param id path string true "Extracurricular ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /extracurriculars/{id} [delete]
func (h *ExtracurricularHandler) DeleteExtracurricular(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID := c.Locals("user_id").(string)
	extracurricularID := c.Params("id")

	if extracurricularID == "" {
		extracurricularHandlerLogger.Error("Missing extracurricular ID for deletion", map[string]interface{}{
			"user_id": userID,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Extracurricular ID is required",
		})
	}

	err := h.service.DeleteExtracurricular(ctx, extracurricularID, userID)
	if err != nil {
		extracurricularHandlerLogger.Error("Failed to delete extracurricular", map[string]interface{}{
			"error":              err.Error(),
			"user_id":            userID,
			"extracurricular_id": extracurricularID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete extracurricular",
		})
	}

	extracurricularHandlerLogger.Info("Extracurricular deleted successfully", map[string]interface{}{
		"extracurricular_id": extracurricularID,
		"user_id":            userID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Extracurricular deleted successfully",
		Data:    nil,
	})
}
