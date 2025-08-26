package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/internal/calendar"
	"github.com/paccy-io/paccy-backend/internal/calendar/dto"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	response "github.com/paccy-io/paccy-backend/pkg/response"
)

var calendarHandlerLogger = logger.New("calendar_handler")

type CalendarHandler struct {
	calendarService calendar.CalendarService
}

func NewCalendarHandler(calendarService calendar.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
	}
}

// CreateEvent creates a new event
// @Summary Create Event
// @Description Create a new calendar event
// @Tags Calendar
// @Accept json
// @Produce json
// @Param event body dto.CreateEventDTO true "Create Event Data"
// @Success 201 {object} response.SuccessResponse{data=dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar [post]
func (h *CalendarHandler) CreateEvent(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var createDTO dto.CreateEventDTO
	if err := c.BodyParser(&createDTO); err != nil {
		calendarHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	event, err := h.calendarService.CreateEvent(c.Context(), userID, &createDTO)
	if err != nil {
		calendarHandlerLogger.Error("Failed to create event", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to create event",
		})
	}

	calendarHandlerLogger.Info("Event created successfully", map[string]interface{}{
		"user_id":  userID,
		"event_id": event.ID,
		"title":    event.Title,
	})

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse{
		Success: true,
		Message: "Event created successfully",
		Data:    event,
	})
}

// GetEvent retrieves a specific event by ID
// @Summary Get Event
// @Description Get a specific calendar event by ID
// @Tags Calendar
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} response.SuccessResponse{data=dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/{id} [get]
func (h *CalendarHandler) GetEvent(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	eventID := c.Params("id")

	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event ID is required",
		})
	}

	event, err := h.calendarService.GetEventByID(c.Context(), userID, eventID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get event", map[string]interface{}{
			"user_id":  userID,
			"event_id": eventID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusNotFound).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Event retrieved successfully",
		Data:    event,
	})
}

// GetUserEvents retrieves all events for the authenticated user
// @Summary Get User Events
// @Description Get all calendar events for the authenticated user
// @Tags Calendar
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar [get]
func (h *CalendarHandler) GetUserEvents(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	events, err := h.calendarService.GetUserEvents(c.Context(), userID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get user events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
	}

	calendarHandlerLogger.Info("User events retrieved successfully", map[string]interface{}{
		"user_id": userID,
		"count":   len(events),
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetInboxEvents retrieves all inbox events (events without dates)
// @Summary Get Inbox Events
// @Description Get all calendar events without dates (inbox events)
// @Tags Calendar
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/inbox [get]
func (h *CalendarHandler) GetInboxEvents(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	events, err := h.calendarService.GetInboxEvents(c.Context(), userID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get inbox events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve inbox events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Inbox events retrieved successfully",
		Data:    events,
	})
}

// GetEventsByDateRange retrieves events within a date range
// @Summary Get Events by Date Range
// @Description Get calendar events within a specific date range
// @Tags Calendar
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/date-range [get]
func (h *CalendarHandler) GetEventsByDateRange(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Both start_date and end_date are required",
		})
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid start_date format. Use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid end_date format. Use YYYY-MM-DD",
		})
	}

	events, err := h.calendarService.GetEventsByDateRange(c.Context(), userID, startDate, endDate)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get events by date range", map[string]interface{}{
			"user_id":    userID,
			"start_date": startDateStr,
			"end_date":   endDateStr,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetEventsByType retrieves events by type
// @Summary Get Events by Type
// @Description Get calendar events by specific type
// @Tags Calendar
// @Produce json
// @Param type query string true "Event type"
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/type [get]
func (h *CalendarHandler) GetEventsByType(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	eventType := c.Query("type")

	if eventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event type is required",
		})
	}

	events, err := h.calendarService.GetEventsByType(c.Context(), userID, eventType)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get events by type", map[string]interface{}{
			"user_id":    userID,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetEventsByPriority retrieves events by priority
// @Summary Get Events by Priority
// @Description Get calendar events by specific priority
// @Tags Calendar
// @Produce json
// @Param priority query string true "Event priority"
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/priority [get]
func (h *CalendarHandler) GetEventsByPriority(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	priority := c.Query("priority")

	if priority == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event priority is required",
		})
	}

	events, err := h.calendarService.GetEventsByPriority(c.Context(), userID, priority)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get events by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}

// GetUpcomingEvents retrieves upcoming events within specified days
// @Summary Get Upcoming Events
// @Description Get calendar events coming up within specified days
// @Tags Calendar
// @Produce json
// @Param days query int false "Number of days to look ahead (default: 7)"
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/upcoming [get]
func (h *CalendarHandler) GetUpcomingEvents(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	daysStr := c.Query("days", "7") // Default to 7 days

	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid days parameter. Must be a positive integer",
		})
	}

	events, err := h.calendarService.GetUpcomingEvents(c.Context(), userID, days)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get upcoming events", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve upcoming events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Upcoming events retrieved successfully",
		Data:    events,
	})
}

// UpdateEvent updates an existing event
// @Summary Update Event
// @Description Update an existing calendar event
// @Tags Calendar
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param event body dto.UpdateEventDTO true "Update Event Data"
// @Success 200 {object} response.SuccessResponse{data=dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/{id} [put]
func (h *CalendarHandler) UpdateEvent(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	eventID := c.Params("id")

	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event ID is required",
		})
	}

	var updateDTO dto.UpdateEventDTO
	if err := c.BodyParser(&updateDTO); err != nil {
		calendarHandlerLogger.Error("Failed to parse request body", map[string]interface{}{
			"user_id":  userID,
			"event_id": eventID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	event, err := h.calendarService.UpdateEvent(c.Context(), userID, eventID, &updateDTO)
	if err != nil {
		calendarHandlerLogger.Error("Failed to update event", map[string]interface{}{
			"user_id":  userID,
			"event_id": eventID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to update event",
		})
	}

	calendarHandlerLogger.Info("Event updated successfully", map[string]interface{}{
		"user_id":  userID,
		"event_id": eventID,
		"title":    event.Title,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Event updated successfully",
		Data:    event,
	})
}

// DeleteEvent deletes an event
// @Summary Delete Event
// @Description Delete a calendar event
// @Tags Calendar
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/{id} [delete]
func (h *CalendarHandler) DeleteEvent(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	eventID := c.Params("id")

	if eventID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Event ID is required",
		})
	}

	err := h.calendarService.DeleteEvent(c.Context(), userID, eventID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to delete event", map[string]interface{}{
			"user_id":  userID,
			"event_id": eventID,
			"error":    err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to delete event",
		})
	}

	calendarHandlerLogger.Info("Event deleted successfully", map[string]interface{}{
		"user_id":  userID,
		"event_id": eventID,
	})

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Event deleted successfully",
		Data:    nil,
	})
}

// GetEventCount retrieves the total count of events for the user
// @Summary Get Event Count
// @Description Get the total number of events for the authenticated user
// @Tags Calendar
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=map[string]int64}
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/calendar/count [get]
func (h *CalendarHandler) GetEventCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	count, err := h.calendarService.GetEventCount(c.Context(), userID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get event count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve event count",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Event count retrieved successfully",
		Data: map[string]int64{
			"count": count,
		},
	})
}

// GetEventsByCollegeApplication retrieves events by college application ID
// @Summary Get Events by College Application
// @Description Get calendar events associated with a specific college application
// @Tags Calendar
// @Produce json
// @Param college_application_id path string true "College Application ID"
// @Success 200 {object} response.SuccessResponse{data=[]dto.EventResponseDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/college-applications/{college_application_id}/calendar [get]
func (h *CalendarHandler) GetEventsByCollegeApplication(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collegeApplicationID := c.Params("college_application_id")

	if collegeApplicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Success: false,
			Error:   "College application ID is required",
		})
	}

	events, err := h.calendarService.GetEventsByCollegeApplication(c.Context(), userID, collegeApplicationID)
	if err != nil {
		calendarHandlerLogger.Error("Failed to get events by college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessResponse{
		Success: true,
		Message: "Events retrieved successfully",
		Data:    events,
	})
}
