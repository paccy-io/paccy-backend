package handlers

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/academic"
	"github.com/paccy-io/paccy-backend/internal/academic/dto"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	response "github.com/paccy-io/paccy-backend/pkg/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var handlerLogger = logger.New("academic_handler")

type AcademicHandler struct {
	service   academic.AcademicService
	validator *validator.Validate
}

func NewAcademicHandler(service academic.AcademicService) *AcademicHandler {
	return &AcademicHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Exam handlers

// CreateExam creates a new exam
// @Summary Create a new exam
// @Description Create a new exam record for the authenticated user
// @Tags Academic
// @Accept json
// @Produce json
// @Param exam body dto.CreateExamDTO true "Exam data"
// @Success 201 {object} response.SuccessResponse{data=dto.ExamResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams [post]
// @Security BearerAuth
func (h *AcademicHandler) CreateExam(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		handlerLogger.Error("Invalid user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	var req dto.CreateExamDTO
	if err := c.BodyParser(&req); err != nil {
		handlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		handlerLogger.Error("Validation failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	exam, err := h.service.CreateExam(ctx, userID, &req)
	if err != nil {
		handlerLogger.Error("Failed to create exam", map[string]interface{}{
			"error":     err.Error(),
			"exam_name": req.Name,
			"user_id":   userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create exam",
		})
	}

	handlerLogger.Info("Exam created successfully", map[string]interface{}{
		"exam_id":   exam.ID,
		"exam_name": exam.Name,
		"user_id":   userIDStr,
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exam created successfully",
		Data:    exam,
	})
}

// GetExamByID gets an exam by ID
// @Summary Get exam by ID
// @Description Get a specific exam by its ID (only accessible by the owner)
// @Tags Academic
// @Produce json
// @Param id path string true "Exam ID"
// @Success 200 {object} response.SuccessResponse{data=dto.ExamResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams/{id} [get]
// @Security BearerAuth
func (h *AcademicHandler) GetExamByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	examID := c.Params("id")

	exam, err := h.service.GetExamByID(ctx, examID, userIDStr)
	if err != nil {
		handlerLogger.Error("Failed to get exam by ID", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": examID,
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Exam not found",
		})
	}

	handlerLogger.Info("Retrieved exam by ID", map[string]interface{}{
		"exam_id":   examID,
		"exam_name": exam.Name,
		"user_id":   userIDStr,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exam retrieved successfully",
		Data:    exam,
	})
}

// GetExamsByUserID gets all exams for the authenticated user
// @Summary Get user's exams
// @Description Get all exams for the authenticated user
// @Tags Academic
// @Produce json
// @Success 200 {array} dto.ExamResponseDTO
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams [get]
// @Security BearerAuth
func (h *AcademicHandler) GetExamsByUserID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)

	exams, err := h.service.GetExamsByUserID(ctx, userIDStr)
	if err != nil {
		handlerLogger.Error("Failed to get exams by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get exams",
		})
	}

	handlerLogger.Info("Retrieved exams for user", map[string]interface{}{
		"user_id": userIDStr,
		"count":   len(exams),
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exams retrieved successfully",
		Data:    exams,
	})
}

// GetExamsByType gets exams by type for the authenticated user
// @Summary Get exams by type
// @Description Get all exams of a specific type for the authenticated user
// @Tags Academic
// @Produce json
// @Param type path string true "Exam Type" Enums(SAT, ACT, SAT Subject Test, AP, IB, TOEFL, IELTS, GRE, GMAT, LSAT, MCAT, other)
// @Success 200 {array} dto.ExamResponseDTO
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams/type/{type} [get]
// @Security BearerAuth
func (h *AcademicHandler) GetExamsByType(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	examType := c.Params("type")

	exams, err := h.service.GetExamsByType(ctx, userIDStr, examType)
	if err != nil {
		handlerLogger.Error("Failed to get exams by type", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userIDStr,
			"exam_type": examType,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid exam type or failed to get exams",
		})
	}

	handlerLogger.Info("Retrieved exams by type for user", map[string]interface{}{
		"user_id":   userIDStr,
		"exam_type": examType,
		"count":     len(exams),
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exams retrieved successfully",
		Data:    exams,
	})
}

// UpdateExam updates an exam
// @Summary Update exam
// @Description Update an existing exam (only accessible by the owner)
// @Tags Academic
// @Accept json
// @Produce json
// @Param id path string true "Exam ID"
// @Param exam body dto.UpdateExamDTO true "Updated exam data"
// @Success 200 {object} response.SuccessResponse{data=dto.ExamResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams/{id} [put]
// @Security BearerAuth
func (h *AcademicHandler) UpdateExam(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	examID := c.Params("id")

	var req dto.UpdateExamDTO
	if err := c.BodyParser(&req); err != nil {
		handlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": examID,
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		handlerLogger.Error("Update validation failed", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": examID,
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	exam, err := h.service.UpdateExam(ctx, examID, userIDStr, &req)
	if err != nil {
		handlerLogger.Error("Failed to update exam", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": examID,
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update exam",
		})
	}

	handlerLogger.Info("Exam updated successfully", map[string]interface{}{
		"exam_id":   examID,
		"exam_name": exam.Name,
		"user_id":   userIDStr,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exam updated successfully",
		Data:    exam,
	})
}

// DeleteExam deletes an exam
// @Summary Delete exam
// @Description Delete an existing exam (only accessible by the owner)
// @Tags Academic
// @Produce json
// @Param id path string true "Exam ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/exams/{id} [delete]
// @Security BearerAuth
func (h *AcademicHandler) DeleteExam(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	examID := c.Params("id")

	err := h.service.DeleteExam(ctx, examID, userIDStr)
	if err != nil {
		handlerLogger.Error("Failed to delete exam", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": examID,
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete exam",
		})
	}

	handlerLogger.Info("Exam deleted successfully", map[string]interface{}{
		"exam_id": examID,
		"user_id": userIDStr,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Exam deleted successfully",
	})
}

// HighSchoolGPA handlers

// CreateHighSchoolGPA creates a new HighSchoolGPA
// @Summary Create HighSchoolGPA
// @Description Create a new HighSchoolGPA record for the authenticated user (max 1 per user)
// @Tags Academic
// @Accept json
// @Produce json
// @Param gpa body dto.CreateHighSchoolGPADTO true "HighSchoolGPA data"
// @Success 201 {object} response.SuccessResponse{data=dto.HighSchoolGPAResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/high-school-gpa [post]
// @Security BearerAuth
func (h *AcademicHandler) CreateHighSchoolGPA(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		handlerLogger.Error("Invalid user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	var req dto.CreateHighSchoolGPADTO
	if err := c.BodyParser(&req); err != nil {
		handlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		handlerLogger.Error("Validation failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	gpa, err := h.service.CreateHighSchoolGPA(ctx, userID, &req)
	if err != nil {
		handlerLogger.Error("Failed to create HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		// Check if it's a duplicate error
		if err.Error() == "user already exists" {
			return c.Status(fiber.StatusConflict).JSON(response.ErrorResponse{
				Success: false,
				Error:   "HighSchoolGPA already exists for this user",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create HighSchoolGPA",
		})
	}

	handlerLogger.Info("HighSchoolGPA created successfully", map[string]interface{}{
		"gpa_id":  gpa.ID,
		"user_id": userIDStr,
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "HighSchoolGPA created successfully",
		Data:    gpa,
	})
}

// GetHighSchoolGPAByUserID gets the HighSchoolGPA for the authenticated user
// @Summary Get user's HighSchoolGPA
// @Description Get the HighSchoolGPA record for the authenticated user
// @Tags Academic
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=dto.HighSchoolGPAResponseDTO}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/high-school-gpa [get]
// @Security BearerAuth
func (h *AcademicHandler) GetHighSchoolGPAByUserID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)

	gpa, err := h.service.GetHighSchoolGPAByUserID(ctx, userIDStr)
	if err != nil {
		handlerLogger.Error("Failed to get HighSchoolGPA by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
			Success: false,
			Error:   "HighSchoolGPA not found",
		})
	}

	handlerLogger.Info("Retrieved HighSchoolGPA for user", map[string]interface{}{
		"user_id": userIDStr,
		"gpa_id":  gpa.ID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "HighSchoolGPA retrieved successfully",
		Data:    gpa,
	})
}

// UpdateHighSchoolGPA updates the HighSchoolGPA
// @Summary Update HighSchoolGPA
// @Description Update the HighSchoolGPA record for the authenticated user
// @Tags Academic
// @Accept json
// @Produce json
// @Param gpa body dto.UpdateHighSchoolGPADTO true "Updated HighSchoolGPA data"
// @Success 200 {object} response.SuccessResponse{data=dto.HighSchoolGPAResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/high-school-gpa [put]
// @Security BearerAuth
func (h *AcademicHandler) UpdateHighSchoolGPA(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)

	var req dto.UpdateHighSchoolGPADTO
	if err := c.BodyParser(&req); err != nil {
		handlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		handlerLogger.Error("Update validation failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
		})
	}

	gpa, err := h.service.UpdateHighSchoolGPA(ctx, userIDStr, &req)
	if err != nil {
		handlerLogger.Error("Failed to update HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update HighSchoolGPA",
		})
	}

	handlerLogger.Info("HighSchoolGPA updated successfully", map[string]interface{}{
		"gpa_id":  gpa.ID,
		"user_id": userIDStr,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "HighSchoolGPA updated successfully",
		Data:    gpa,
	})
}

// DeleteHighSchoolGPA deletes the HighSchoolGPA
// @Summary Delete HighSchoolGPA
// @Description Delete the HighSchoolGPA record for the authenticated user
// @Tags Academic
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/high-school-gpa [delete]
// @Security BearerAuth
func (h *AcademicHandler) DeleteHighSchoolGPA(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDStr := c.Locals("user_id").(string)

	err := h.service.DeleteHighSchoolGPA(ctx, userIDStr)
	if err != nil {
		handlerLogger.Error("Failed to delete HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete HighSchoolGPA",
		})
	}

	handlerLogger.Info("HighSchoolGPA deleted successfully", map[string]interface{}{
		"user_id": userIDStr,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "HighSchoolGPA deleted successfully",
	})
}
