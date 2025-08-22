package models

import (
	"time"

	"github.com/paccy-io/paccy-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Essay struct {
	models.BaseModel `bson:",inline"` // make them in root level not in BaseModel, by using an inline struct

	Title                string             `bson:"title" json:"title"`
	CollegeApplicationID primitive.ObjectID `bson:"college_application_id" json:"college_application_id"`// ID of the college application in the collection of "college_applications" in MongoDB
	Status               EssayStatus        `bson:"status" json:"status"` // enum: ["not_started", "in_progress", "not_submitted", "review_needed", "submitted", "completed"]
	Prompt               string             `bson:"prompt" json:"prompt"`
	Content              string             `bson:"content" json:"content"`
	AIScore              float64            `bson:"ai_score" json:"ai_score"` // on 10 scale
	AIFeedbacks          []string           `bson:"ai_feedbacks" json:"ai_feedbacks"`
	AISuggestion         string             `bson:"ai_suggestion" json:"ai_suggestion"`
}

type EssayStatus string

const (
	EssayStatusNotStarted   EssayStatus = "not_started"
	EssayStatusInProgress   EssayStatus = "in_progress"
	EssayStatusNotSubmitted EssayStatus = "not_submitted"
	EssayStatusReviewNeeded EssayStatus = "review_needed"
	EssayStatusSubmitted    EssayStatus = "submitted"
	EssayStatusCompleted    EssayStatus = "completed"
)

func GetCollectionName() string {
	return "essays"
}

func (e *Essay) BeforeCreate() {
	e.BaseModel.BeforeCreate()

	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
}

func (e *Essay) BeforeUpdate() {
	e.BaseModel.BeforeUpdate()
}

func IsValidEssayStatus(status string) EssayStatus {
	switch EssayStatus(status) {
	case EssayStatusNotStarted, EssayStatusInProgress, EssayStatusNotSubmitted, EssayStatusReviewNeeded, EssayStatusSubmitted, EssayStatusCompleted:
		return EssayStatus(status)
	default:
		return ""
	}
}

