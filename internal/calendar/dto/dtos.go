package dto

import (
	"encoding/json"
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

// UpdateEventDTO with custom date handling
type UpdateEventDTO struct {
	Title       *string              `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string              `json:"description" validate:"omitempty,max=1000"`
	Date        *time.Time           `json:"date,omitempty"` // can be set to null to move to inbox
	Type        *models.EventType    `json:"type" validate:"omitempty,oneof=application_deadline interview campus_visit test_date reminder other"`
	Priority    *models.TaskPriority `json:"priority" validate:"omitempty,oneof=high medium low"`
	Location    *string              `json:"location" validate:"omitempty,max=255"`
	CollegeID   *string              `json:"college_id" validate:"omitempty"` // can be set to null to remove college reference

	// Internal fields to track what was provided in JSON
	dateProvided bool
}

// Custom UnmarshalJSON to detect when date field is explicitly provided
func (dto *UpdateEventDTO) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to detect which fields are present
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if date field was provided (even if null)
	if _, exists := raw["date"]; exists {
		dto.dateProvided = true
	}

	// Create a type alias to avoid infinite recursion
	type Alias UpdateEventDTO
	aux := (*Alias)(dto)

	// Unmarshal normally
	return json.Unmarshal(data, aux)
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

// ToUpdateMap converts UpdateEventDTO to update operations for database
func (dto *UpdateEventDTO) ToUpdateMap() (map[string]interface{}, map[string]interface{}, error) {
	setUpdates := make(map[string]interface{})
	unsetUpdates := make(map[string]interface{})

	if dto.Title != nil {
		setUpdates["title"] = *dto.Title
	}
	if dto.Description != nil {
		setUpdates["description"] = *dto.Description
	}

	// Check if date field was explicitly provided in JSON
	if dto.dateProvided {
		if dto.Date == nil {
			// Setting date to null means moving to inbox - unset the date field
			unsetUpdates["date"] = ""
		} else {
			// Setting date to a specific value
			setUpdates["date"] = *dto.Date
		}
	}

	if dto.Type != nil {
		setUpdates["type"] = *dto.Type
	}
	if dto.Priority != nil {
		setUpdates["priority"] = *dto.Priority
	}
	if dto.Location != nil {
		setUpdates["location"] = *dto.Location
	}
	if dto.CollegeID != nil {
		if *dto.CollegeID == "" {
			// Remove college reference
			unsetUpdates["college_id"] = ""
		} else {
			collegeObjectID, err := primitive.ObjectIDFromHex(*dto.CollegeID)
			if err != nil {
				return nil, nil, err
			}
			setUpdates["college_id"] = collegeObjectID
		}
	}

	return setUpdates, unsetUpdates, nil
}
