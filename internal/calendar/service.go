package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/calendar/dto"
	"github.com/paccy-io/paccy-backend/internal/calendar/models"
	"github.com/paccy-io/paccy-backend/internal/college"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var serviceLogger = logger.New("calendar_service")

type CalendarService interface {
	CreateEvent(ctx context.Context, userID string, createDTO *dto.CreateEventDTO) (*dto.EventResponseDTO, error)
	GetEventByID(ctx context.Context, userID, eventID string) (*dto.EventResponseDTO, error)
	GetUserEvents(ctx context.Context, userID string) ([]*dto.EventResponseDTO, error)
	GetInboxEvents(ctx context.Context, userID string) ([]*dto.EventResponseDTO, error)
	GetEventsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*dto.EventResponseDTO, error)
	GetEventsByType(ctx context.Context, userID string, eventType string) ([]*dto.EventResponseDTO, error)
	GetEventsByPriority(ctx context.Context, userID string, priority string) ([]*dto.EventResponseDTO, error)
	GetUpcomingEvents(ctx context.Context, userID string, days int) ([]*dto.EventResponseDTO, error)
	GetEventsByCollegeApplication(ctx context.Context, userID, collegeApplicationID string) ([]*dto.EventResponseDTO, error)
	UpdateEvent(ctx context.Context, userID, eventID string, updateDTO *dto.UpdateEventDTO) (*dto.EventResponseDTO, error)
	DeleteEvent(ctx context.Context, userID, eventID string) error
	GetEventCount(ctx context.Context, userID string) (int64, error)
}

type calendarService struct {
	calendarRepo CalendarRepository
	collegeRepo  college.CollegeRepository
	validator    *validator.Validate
}

func NewCalendarService(calendarRepo CalendarRepository, collegeRepo college.CollegeRepository, validator *validator.Validate) CalendarService {
	return &calendarService{
		calendarRepo: calendarRepo,
		collegeRepo:  collegeRepo,
		validator:    validator,
	}
}

func (s *calendarService) CreateEvent(ctx context.Context, userID string, createDTO *dto.CreateEventDTO) (*dto.EventResponseDTO, error) {
	serviceLogger.Info("Creating new event", map[string]interface{}{
		"user_id": userID,
		"title":   createDTO.Title,
		"type":    createDTO.Type,
	})

	// Validate DTO
	if err := s.validator.Struct(createDTO); err != nil {
		serviceLogger.Error("Validation failed for create event", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	// Validate college_id if provided
	if createDTO.CollegeID != nil && *createDTO.CollegeID != "" {
		collegeApp, err := s.collegeRepo.GetByID(ctx, *createDTO.CollegeID)
		if err != nil {
			serviceLogger.Error("College application not found", map[string]interface{}{
				"user_id":    userID,
				"college_id": *createDTO.CollegeID,
				"error":      err.Error(),
			})
			return nil, errors.ErrInvalidInput
		}

		// Verify that the college application belongs to the user
		if collegeApp.UserID.Hex() != userID {
			serviceLogger.Warn("User attempted to create event with college_id they don't own", map[string]interface{}{
				"user_id":       userID,
				"college_id":    *createDTO.CollegeID,
				"college_owner": collegeApp.UserID.Hex(),
			})
			return nil, errors.ErrInvalidInput
		}
	}

	// Convert user ID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		serviceLogger.Error("Invalid user ObjectID", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	// Convert DTO to model
	event, err := createDTO.ToModel(userObjectID)
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to model", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	// Create event
	createdEvent, err := s.calendarRepo.Create(ctx, event)
	if err != nil {
		serviceLogger.Error("Failed to create event", map[string]interface{}{
			"user_id": userID,
			"title":   createDTO.Title,
			"error":   err.Error(),
		})
		return nil, err
	}

	// Convert to response DTO
	responseDTO, err := s.convertToResponseDTO(ctx, createdEvent)
	if err != nil {
		serviceLogger.Error("Failed to convert event to response DTO", map[string]interface{}{
			"event_id": createdEvent.ID.Hex(),
			"error":    err.Error(),
		})
		return nil, err
	}

	serviceLogger.Info("Event created successfully", map[string]interface{}{
		"event_id": createdEvent.ID.Hex(),
		"user_id":  userID,
		"title":    createdEvent.Title,
	})

	return responseDTO, nil
}

func (s *calendarService) GetEventByID(ctx context.Context, userID, eventID string) (*dto.EventResponseDTO, error) {
	event, err := s.calendarRepo.GetByID(ctx, eventID)
	if err != nil {
		serviceLogger.Error("Failed to get event by ID", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Verify ownership
	if event.UserID.Hex() != userID {
		serviceLogger.Warn("User attempted to access event they don't own", map[string]interface{}{
			"event_id":        eventID,
			"requesting_user": userID,
			"event_owner":     event.UserID.Hex(),
		})
		return nil, errors.ErrGetFailed
	}

	responseDTO, err := s.convertToResponseDTO(ctx, event)
	if err != nil {
		serviceLogger.Error("Failed to convert event to response DTO", map[string]interface{}{
			"event_id": eventID,
			"error":    err.Error(),
		})
		return nil, err
	}

	return responseDTO, nil
}

func (s *calendarService) GetUserEvents(ctx context.Context, userID string) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get user events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	serviceLogger.Info("Retrieved user events", map[string]interface{}{
		"user_id": userID,
		"count":   len(responseDTOs),
	})

	return responseDTOs, nil
}

func (s *calendarService) GetInboxEvents(ctx context.Context, userID string) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetInboxEvents(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get inbox events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert inbox events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	serviceLogger.Info("Retrieved inbox events", map[string]interface{}{
		"user_id": userID,
		"count":   len(responseDTOs),
	})

	return responseDTOs, nil
}

func (s *calendarService) GetEventsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetEventsByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		serviceLogger.Error("Failed to get events by date range", map[string]interface{}{
			"user_id":    userID,
			"start_date": startDate,
			"end_date":   endDate,
			"error":      err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return responseDTOs, nil
}

func (s *calendarService) GetEventsByType(ctx context.Context, userID string, eventType string) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetEventsByType(ctx, userID, eventType)
	if err != nil {
		serviceLogger.Error("Failed to get events by type", map[string]interface{}{
			"user_id":    userID,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return responseDTOs, nil
}

func (s *calendarService) GetEventsByPriority(ctx context.Context, userID string, priority string) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetEventsByPriority(ctx, userID, priority)
	if err != nil {
		serviceLogger.Error("Failed to get events by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return responseDTOs, nil
}

func (s *calendarService) GetUpcomingEvents(ctx context.Context, userID string, days int) ([]*dto.EventResponseDTO, error) {
	events, err := s.calendarRepo.GetUpcomingEvents(ctx, userID, days)
	if err != nil {
		serviceLogger.Error("Failed to get upcoming events", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert upcoming events to response DTOs", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return responseDTOs, nil
}

func (s *calendarService) UpdateEvent(ctx context.Context, userID, eventID string, updateDTO *dto.UpdateEventDTO) (*dto.EventResponseDTO, error) {
	serviceLogger.Info("Updating event", map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	})

	// Validate DTO
	if err := s.validator.Struct(updateDTO); err != nil {
		serviceLogger.Error("Validation failed for update event", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Validate college_id if provided in update
	if updateDTO.CollegeID != nil {
		if *updateDTO.CollegeID != "" {
			// If college_id is not empty, validate it exists and belongs to user
			collegeApp, err := s.collegeRepo.GetByID(ctx, *updateDTO.CollegeID)
			if err != nil {
				serviceLogger.Error("College application not found during update", map[string]interface{}{
					"user_id":    userID,
					"event_id":   eventID,
					"college_id": *updateDTO.CollegeID,
					"error":      err.Error(),
				})
				return nil, errors.ErrInvalidInput
			}

			// Verify that the college application belongs to the user
			if collegeApp.UserID.Hex() != userID {
				serviceLogger.Warn("User attempted to update event with college_id they don't own", map[string]interface{}{
					"user_id":       userID,
					"event_id":      eventID,
					"college_id":    *updateDTO.CollegeID,
					"college_owner": collegeApp.UserID.Hex(),
				})
				return nil, errors.ErrInvalidInput
			}
		}
		// If college_id is empty string, it means user wants to remove the college reference, which is valid
	}

	// Check if event exists and user owns it
	existingEvent, err := s.calendarRepo.GetByID(ctx, eventID)
	if err != nil {
		serviceLogger.Error("Failed to get existing event for update", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Verify ownership
	if existingEvent.UserID.Hex() != userID {
		serviceLogger.Warn("User attempted to update event they don't own", map[string]interface{}{
			"event_id":        eventID,
			"requesting_user": userID,
			"event_owner":     existingEvent.UserID.Hex(),
		})
		return nil, errors.ErrUpdateFailed
	}

	// Convert DTO to update map
	updates, err := updateDTO.ToUpdateMap()
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to update map", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Update event
	updatedEvent, err := s.calendarRepo.Update(ctx, eventID, updates)
	if err != nil {
		serviceLogger.Error("Failed to update event", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Convert to response DTO
	responseDTO, err := s.convertToResponseDTO(ctx, updatedEvent)
	if err != nil {
		serviceLogger.Error("Failed to convert updated event to response DTO", map[string]interface{}{
			"event_id": eventID,
			"error":    err.Error(),
		})
		return nil, err
	}

	serviceLogger.Info("Event updated successfully", map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
		"title":    updatedEvent.Title,
	})

	return responseDTO, nil
}

func (s *calendarService) DeleteEvent(ctx context.Context, userID, eventID string) error {
	serviceLogger.Info("Deleting event", map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	})

	// Check if event exists and user owns it
	existingEvent, err := s.calendarRepo.GetByID(ctx, eventID)
	if err != nil {
		serviceLogger.Error("Failed to get existing event for deletion", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return err
	}

	// Verify ownership
	if existingEvent.UserID.Hex() != userID {
		serviceLogger.Warn("User attempted to delete event they don't own", map[string]interface{}{
			"event_id":        eventID,
			"requesting_user": userID,
			"event_owner":     existingEvent.UserID.Hex(),
		})
		return errors.ErrDeleteFailed
	}

	// Delete event
	err = s.calendarRepo.Delete(ctx, eventID)
	if err != nil {
		serviceLogger.Error("Failed to delete event", map[string]interface{}{
			"event_id": eventID,
			"user_id":  userID,
			"error":    err.Error(),
		})
		return err
	}

	serviceLogger.Info("Event deleted successfully", map[string]interface{}{
		"event_id": eventID,
		"user_id":  userID,
	})

	return nil
}

func (s *calendarService) GetEventCount(ctx context.Context, userID string) (int64, error) {
	count, err := s.calendarRepo.Count(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get event count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, err
	}

	return count, nil
}

func (s *calendarService) GetEventsByCollegeApplication(ctx context.Context, userID, collegeApplicationID string) ([]*dto.EventResponseDTO, error) {
	serviceLogger.Info("Getting events by college application", map[string]interface{}{
		"user_id":                userID,
		"college_application_id": collegeApplicationID,
	})

	// Validate college application ID format
	_, err := primitive.ObjectIDFromHex(collegeApplicationID)
	if err != nil {
		serviceLogger.Error("Invalid college application ObjectID", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	// Verify that the college application exists and belongs to the user
	collegeApp, err := s.collegeRepo.GetByID(ctx, collegeApplicationID)
	if err != nil {
		serviceLogger.Error("College application not found", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify ownership of the college application
	if collegeApp.UserID.Hex() != userID {
		serviceLogger.Warn("User attempted to access events for college application they don't own", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"college_owner":          collegeApp.UserID.Hex(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Get events by college application ID
	events, err := s.calendarRepo.GetEventsByCollegeApplication(ctx, userID, collegeApplicationID)
	if err != nil {
		serviceLogger.Error("Failed to get events by college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, err
	}

	responseDTOs, err := s.convertEventsToResponseDTOs(ctx, events)
	if err != nil {
		serviceLogger.Error("Failed to convert events to response DTOs", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, err
	}

	serviceLogger.Info("Retrieved events by college application", map[string]interface{}{
		"user_id":                userID,
		"college_application_id": collegeApplicationID,
		"count":                  len(responseDTOs),
	})

	return responseDTOs, nil
}

// Helper methods
func (s *calendarService) convertToResponseDTO(ctx context.Context, event *models.Event) (*dto.EventResponseDTO, error) {
	var collegeName *string

	// Resolve college name if college_id is present
	if event.CollegeID != nil && !event.CollegeID.IsZero() {
		college, err := s.collegeRepo.GetByID(ctx, event.CollegeID.Hex())
		if err != nil {
			serviceLogger.Warn("Failed to resolve college name for event", map[string]interface{}{
				"event_id":   event.ID.Hex(),
				"college_id": event.CollegeID.Hex(),
				"error":      err.Error(),
			})
			// Don't fail the entire operation, just set collegeName to nil
		} else {
			collegeName = &college.Name
		}
	}

	return dto.EventToResponseDTO(event, collegeName), nil
}

func (s *calendarService) convertEventsToResponseDTOs(ctx context.Context, events []*models.Event) ([]*dto.EventResponseDTO, error) {
	responseDTOs := make([]*dto.EventResponseDTO, len(events))

	for i, event := range events {
		responseDTO, err := s.convertToResponseDTO(ctx, event)
		if err != nil {
			serviceLogger.Error("Failed to convert event to response DTO", map[string]interface{}{
				"event_id": event.ID.Hex(),
				"error":    err.Error(),
			})
			return nil, fmt.Errorf("failed to convert event %s: %w", event.ID.Hex(), err)
		}
		responseDTOs[i] = responseDTO
	}

	return responseDTOs, nil
}
