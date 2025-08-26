package dto

import (
	"time"

	"github.com/paccy-io/paccy-backend/internal/calendar/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateEventDTO struct {
	Title       string              `json:"title" validate:"required,min=1,max=255"`
	Description string              `json:"description" validate:"max=1000"`
	Date        *time.Time          `json:"date"` // optional, null for inbox events
	Type        models.EventType    `json:"type" validate:"required,oneof=application_deadline interview campus_visit test_date reminder other"`
	Priority    models.TaskPriority `json:"priority" validate:"required,oneof=high medium low"`
	Location    string              `json:"location" validate:"max=255"`
	CollegeID   *string             `json:"college_id" validate:"omitempty"` // optional college reference
}

type UpdateEventDTO struct {
	Title       *string              `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string              `json:"description" validate:"omitempty,max=1000"`
	Date        *time.Time           `json:"date"` // can be set to null to move to inbox
	Type        *models.EventType    `json:"type" validate:"omitempty,oneof=application_deadline interview campus_visit test_date reminder other"`
	Priority    *models.TaskPriority `json:"priority" validate:"omitempty,oneof=high medium low"`
	Location    *string              `json:"location" validate:"omitempty,max=255"`
	CollegeID   *string              `json:"college_id" validate:"omitempty"` // can be set to null to remove college reference
}

// EventResponseDTO includes college name for frontend display
type EventResponseDTO struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Date        *time.Time          `json:"date"` // null for inbox events
	Type        models.EventType    `json:"type"`
	Priority    models.TaskPriority `json:"priority"`
	Location    string              `json:"location"`
	CollegeID   *string             `json:"college_id"`   // null if no college reference
	CollegeName *string             `json:"college_name"` // null if no college reference, populated from college application
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

// Conversion methods
func (dto *CreateEventDTO) ToModel(userID primitive.ObjectID) (*models.Event, error) {
	event := &models.Event{
		UserID:      userID,
		Title:       dto.Title,
		Description: dto.Description,
		Date:        dto.Date,
		Type:        dto.Type,
		Priority:    dto.Priority,
		Location:    dto.Location,
	}

	// Handle optional college ID
	if dto.CollegeID != nil && *dto.CollegeID != "" {
		collegeObjectID, err := primitive.ObjectIDFromHex(*dto.CollegeID)
		if err != nil {
			return nil, err
		}
		event.CollegeID = &collegeObjectID
	}

	event.BeforeCreate()
	return event, nil
}

func EventToResponseDTO(event *models.Event, collegeName *string) *EventResponseDTO {
	var collegeID *string
	if event.CollegeID != nil {
		id := event.CollegeID.Hex()
		collegeID = &id
	}

	return &EventResponseDTO{
		ID:          event.ID.Hex(),
		Title:       event.Title,
		Description: event.Description,
		Date:        event.Date,
		Type:        event.Type,
		Priority:    event.Priority,
		Location:    event.Location,
		CollegeID:   collegeID,
		CollegeName: collegeName,
		CreatedAt:   event.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   event.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToUpdateMap converts UpdateEventDTO to a bson.M for database updates
func (dto *UpdateEventDTO) ToUpdateMap() (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if dto.Title != nil {
		updates["title"] = *dto.Title
	}
	if dto.Description != nil {
		updates["description"] = *dto.Description
	}
	if dto.Date != nil {
		updates["date"] = *dto.Date
	}
	if dto.Type != nil {
		updates["type"] = *dto.Type
	}
	if dto.Priority != nil {
		updates["priority"] = *dto.Priority
	}
	if dto.Location != nil {
		updates["location"] = *dto.Location
	}
	if dto.CollegeID != nil {
		if *dto.CollegeID == "" {
			// Set to nil to remove college reference
			updates["college_id"] = nil
		} else {
			collegeObjectID, err := primitive.ObjectIDFromHex(*dto.CollegeID)
			if err != nil {
				return nil, err
			}
			updates["college_id"] = collegeObjectID
		}
	}

	return updates, nil
}
