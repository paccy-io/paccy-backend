package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/college"
	"github.com/paccy-io/paccy-backend/internal/college/dto"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	response "github.com/paccy-io/paccy-backend/pkg/response"
)

var collegeHandlerLogger = logger.New("college-handler")

type CollegeHandler struct {
	collegeService college.CollegeService
}

func NewCollegeHandler(collegeService college.CollegeService) *CollegeHandler {
	return &CollegeHandler{
		collegeService: collegeService,
	}
}

// CreateCollegeApplication creates a new college application for the authenticated user
func (h *CollegeHandler) CreateCollegeApplication(c *fiber.Ctx) error {
	// Get user ID from JWT claims (set by auth middleware)
	userID := c.Locals("user_id").(string)

	var createDTO dto.CreateCollegeApplicationDTO
	if err := c.BodyParser(&createDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
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

	application, err := h.collegeService.Create(ctx, userID, &createDTO)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		collegeHandlerLogger.Error("Failed to create college application", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create college application",
		})
	}

	collegeHandlerLogger.Info("College application created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": application.ID.Hex(),
		"college_name":   application.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "College application created successfully",
		Data:    application,
	})
}

// GetCollegeApplication gets a specific college application by ID
func (h *CollegeHandler) GetCollegeApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("id")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	application, err := h.collegeService.GetByID(ctx, userID, applicationID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "College application not found",
			})
		}
		collegeHandlerLogger.Error("Failed to get college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get college application",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College application retrieved successfully",
		Data:    application,
	})
}

// GetCollegeApplications gets all college applications for the authenticated user
func (h *CollegeHandler) GetCollegeApplications(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	applications, err := h.collegeService.GetByUserID(ctx, userID)
	if err != nil {
		collegeHandlerLogger.Error("Failed to get college applications", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get college applications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College applications retrieved successfully",
		Data:    applications,
	})
}

// UpdateCollegeApplication updates a specific college application
func (h *CollegeHandler) UpdateCollegeApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("id")

	var updateDTO dto.UpdateCollegeApplicationDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	application, err := h.collegeService.Update(ctx, userID, applicationID, &updateDTO)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "College application not found",
			})
		}
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		collegeHandlerLogger.Error("Failed to update college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update college application",
		})
	}

	collegeHandlerLogger.Info("College application updated successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College application updated successfully",
		Data:    application,
	})
}

// DeleteCollegeApplication deletes a specific college application
func (h *CollegeHandler) DeleteCollegeApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("id")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.collegeService.Delete(ctx, userID, applicationID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "College application not found",
			})
		}
		collegeHandlerLogger.Error("Failed to delete college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete college application",
		})
	}

	collegeHandlerLogger.Info("College application deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College application deleted successfully",
		Data:    nil,
	})
}

// GetCollegeApplicationsByStatus gets college applications filtered by status
func (h *CollegeHandler) GetCollegeApplicationsByStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	status := c.Params("status")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	applications, err := h.collegeService.GetByStatus(ctx, userID, status)
	if err != nil {
		collegeHandlerLogger.Error("Failed to get college applications by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get college applications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College applications retrieved successfully",
		Data:    applications,
	})
}

// GetCollegeApplicationsByPriority gets college applications filtered by priority
func (h *CollegeHandler) GetCollegeApplicationsByPriority(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	priority := c.Params("priority")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	applications, err := h.collegeService.GetByPriority(ctx, userID, priority)
	if err != nil {
		collegeHandlerLogger.Error("Failed to get college applications by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get college applications",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "College applications retrieved successfully",
		Data:    applications,
	})
}

// GetUpcomingDeadlines gets college applications with upcoming deadlines
func (h *CollegeHandler) GetUpcomingDeadlines(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Parse days parameter (default to 30 days)
	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 {
		days = 30
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	applications, err := h.collegeService.GetUpcomingDeadlines(ctx, userID, days)
	if err != nil {
		collegeHandlerLogger.Error("Failed to get upcoming deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get upcoming deadlines",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Upcoming deadlines retrieved successfully",
		Data:    applications,
	})
}

// Checklist operations

// CreateChecklist creates a new checklist for a specific college application
func (h *CollegeHandler) CreateChecklist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var createDTO dto.StandaloneCreateChecklistDTO
	if err := c.BodyParser(&createDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
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

	checklist, err := h.collegeService.CreateChecklist(ctx, userID, &createDTO)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "College application not found",
			})
		}
		collegeHandlerLogger.Error("Failed to create checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create checklist",
		})
	}

	collegeHandlerLogger.Info("Checklist created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"checklist_id":   checklist.ID.Hex(),
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "Checklist created successfully",
		Data:    checklist,
	})
}

// GetChecklist gets a specific checklist by ID
func (h *CollegeHandler) GetChecklist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	checklist, err := h.collegeService.GetChecklist(ctx, userID, applicationID, checklistID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Checklist not found",
			})
		}
		collegeHandlerLogger.Error("Failed to get checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get checklist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Checklist retrieved successfully",
		Data:    checklist,
	})
}

// UpdateChecklist updates a specific checklist
func (h *CollegeHandler) UpdateChecklist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")

	var updateDTO dto.StandaloneUpdateChecklistDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	checklist, err := h.collegeService.UpdateChecklist(ctx, userID, applicationID, checklistID, &updateDTO)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Checklist not found",
			})
		}
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		collegeHandlerLogger.Error("Failed to update checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update checklist",
		})
	}

	collegeHandlerLogger.Info("Checklist updated successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Checklist updated successfully",
		Data:    checklist,
	})
}

// DeleteChecklist deletes a specific checklist
func (h *CollegeHandler) DeleteChecklist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.collegeService.DeleteChecklist(ctx, userID, applicationID, checklistID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Checklist not found",
			})
		}
		collegeHandlerLogger.Error("Failed to delete checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete checklist",
		})
	}

	collegeHandlerLogger.Info("Checklist deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Checklist deleted successfully",
		Data:    nil,
	})
}

// Task operations

// CreateTask creates a new task for a specific checklist
func (h *CollegeHandler) CreateTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var createDTO dto.StandaloneCreateTaskDTO
	if err := c.BodyParser(&createDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
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

	task, err := h.collegeService.CreateTask(ctx, userID, &createDTO)
	if err != nil {
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Checklist not found",
			})
		}
		collegeHandlerLogger.Error("Failed to create task", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"checklist_id":   createDTO.ChecklistID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create task",
		})
	}

	collegeHandlerLogger.Info("Task created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"checklist_id":   createDTO.ChecklistID,
		"task_id":        task.ID.Hex(),
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "Task created successfully",
		Data:    task,
	})
}

// GetTask gets a specific task by ID
func (h *CollegeHandler) GetTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")
	taskID := c.Params("taskId")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	task, err := h.collegeService.GetTask(ctx, userID, applicationID, checklistID, taskID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Task not found",
			})
		}
		collegeHandlerLogger.Error("Failed to get task", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to get task",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Task retrieved successfully",
		Data:    task,
	})
}

// UpdateTask updates a specific task
func (h *CollegeHandler) UpdateTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")
	taskID := c.Params("taskId")

	var updateDTO dto.StandaloneUpdateTaskDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		collegeHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	task, err := h.collegeService.UpdateTask(ctx, userID, applicationID, checklistID, taskID, &updateDTO)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Task not found",
			})
		}
		if err == errors.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Invalid input data",
			})
		}
		collegeHandlerLogger.Error("Failed to update task", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update task",
		})
	}

	collegeHandlerLogger.Info("Task updated successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
		"task_id":        taskID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Task updated successfully",
		Data:    task,
	})
}

// DeleteTask deletes a specific task
func (h *CollegeHandler) DeleteTask(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	applicationID := c.Params("applicationId")
	checklistID := c.Params("checklistId")
	taskID := c.Params("taskId")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	err := h.collegeService.DeleteTask(ctx, userID, applicationID, checklistID, taskID)
	if err != nil {
		if err == errors.ErrCollegeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
				Success: false,
				Error:   "Task not found",
			})
		}
		collegeHandlerLogger.Error("Failed to delete task", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete task",
		})
	}

	collegeHandlerLogger.Info("Task deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
		"task_id":        taskID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Task deleted successfully",
		Data:    nil,
	})
}
