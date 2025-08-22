package dto

import (
	"github.com/paccy-io/paccy-backend/internal/essay/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateEssayDTO struct {
	Title                string             `json:"title" validate:"required,min=1,max=255"`
	CollegeApplicationID string             `json:"college_application_id" validate:"required"`
	Status               models.EssayStatus `json:"status" validate:"required,oneof=not_started in_progress not_submitted review_needed submitted completed"`
	Prompt               string             `json:"prompt" validate:"max=2000"`
	Content              string             `json:"content" validate:"max=10000"`
}

type UpdateEssayDTO struct {
	Title        *string             `json:"title" validate:"omitempty,min=1,max=255"`
	Status       *models.EssayStatus `json:"status" validate:"omitempty,oneof=not_started in_progress not_submitted review_needed submitted completed"`
	Prompt       *string             `json:"prompt" validate:"omitempty,max=2000"`
	Content      *string             `json:"content" validate:"omitempty,max=10000"`
	AIScore      *float64            `json:"ai_score" validate:"omitempty,min=0,max=10"`
	AIFeedbacks  []string            `json:"ai_feedbacks" validate:"dive,max=500"`
	AISuggestion *string             `json:"ai_suggestion" validate:"omitempty,max=1000"`
}

// EssayResponseDTO includes college name for frontend display
type EssayResponseDTO struct {
	ID                   string             `json:"id"`
	Title                string             `json:"title"`
	CollegeApplicationID string             `json:"college_application_id"`
	CollegeName          string             `json:"college_name"` // This will be populated from college application
	Status               models.EssayStatus `json:"status"`
	Prompt               string             `json:"prompt"`
	Content              string             `json:"content"`
	AIScore              float64            `json:"ai_score"`
	AIFeedbacks          []string           `json:"ai_feedbacks"`
	AISuggestion         string             `json:"ai_suggestion"`
	CreatedAt            string             `json:"created_at"`
	UpdatedAt            string             `json:"updated_at"`
}

// Conversion methods
func (dto *CreateEssayDTO) ToModel(userID primitive.ObjectID) (*models.Essay, error) {
	collegeAppID, err := primitive.ObjectIDFromHex(dto.CollegeApplicationID)
	if err != nil {
		return nil, err
	}

	essay := &models.Essay{
		Title:                dto.Title,
		CollegeApplicationID: collegeAppID,
		Status:               dto.Status,
		Prompt:               dto.Prompt,
		Content:              dto.Content,
		AIScore:              0.0,
		AIFeedbacks:          []string{},
		AISuggestion:         "",
	}

	essay.BeforeCreate()
	return essay, nil
}

func EssayToResponseDTO(essay *models.Essay, collegeName string) *EssayResponseDTO {
	return &EssayResponseDTO{
		ID:                   essay.ID.Hex(),
		Title:                essay.Title,
		CollegeApplicationID: essay.CollegeApplicationID.Hex(),
		CollegeName:          collegeName,
		Status:               essay.Status,
		Prompt:               essay.Prompt,
		Content:              essay.Content,
		AIScore:              essay.AIScore,
		AIFeedbacks:          essay.AIFeedbacks,
		AISuggestion:         essay.AISuggestion,
		CreatedAt:            essay.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            essay.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
