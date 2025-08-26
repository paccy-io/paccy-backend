package models

import (
	"time"

	"github.com/paccy-io/paccy-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Title            string              `bson:"title" json:"title"`
	Description      string              `bson:"description" json:"description"`
	Date             *time.Time          `bson:"date" json:"date"`         // optional, if no date then it belongs to inbox section
	Type             EventType           `bson:"type" json:"type"`         // "application_deadline", "interview", "campus_visit", "test_date", "reminder", "other"
	Priority         TaskPriority        `bson:"priority" json:"priority"` // "high", "medium", "low"
	Location         string              `bson:"location" json:"location"`
	CollegeID        *primitive.ObjectID `bson:"college_id,omitempty" json:"college_id,omitempty"` //optional but if it exists, it should be a real ObjectID of a college in colleges collection in MongoDB
}

type EventType string

const (
	EventTypeApplicationDeadline EventType = "application_deadline"
	EventTypeInterview           EventType = "interview"
	EventTypeCampusVisit         EventType = "campus_visit"
	EventTypeTestDate            EventType = "test_date"
	EventTypeReminder            EventType = "reminder"
	EventTypeOther               EventType = "other"
)

type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)

func GetCollectionName() string {
	return "events"
}

func (e *Event) BeforeCreate() {
	e.BaseModel.BeforeCreate()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
}

func (e *Event) BeforeUpdate() {
	e.BaseModel.BeforeUpdate()
}

func IsValidEventType(eventType string) EventType {
	switch EventType(eventType) {
	case EventTypeApplicationDeadline, EventTypeInterview, EventTypeCampusVisit, EventTypeTestDate, EventTypeReminder, EventTypeOther:
		return EventType(eventType)
	default:
		return ""
	}
}

func IsValidTaskPriority(priority string) TaskPriority {
	switch TaskPriority(priority) {
	case TaskPriorityHigh, TaskPriorityMedium, TaskPriorityLow:
		return TaskPriority(priority)
	default:
		return ""
	}
}
