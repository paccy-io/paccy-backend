package dto

import (
	"github.com/paccy-io/paccy-backend/internal/extracurriculars/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateExtracurricularDTO struct {
	ActivityName string          `json:"activity_name" validate:"required,min=1,max=255"`
	Category     models.Category `json:"category" validate:"required,oneof=academic sports music arts community_service leadership technology volunteering clubs work_experience research other"`
	Position     string          `json:"position" validate:"required,min=1,max=255"`
	HoursPerWeek int             `json:"hours_per_week" validate:"required,min=1,max=168"`
	StartDate    string          `json:"start_date" validate:"required,len=7"` // YYYY-MM format
	EndDate      string          `json:"end_date" validate:"required"`         // YYYY-MM format or "present"
	Description  string          `json:"description" validate:"max=1000"`
	Location     string          `json:"location" validate:"max=255"`
	Website      string          `json:"website" validate:"omitempty,url,max=255"`
	Achievements []string        `json:"achievements"`
	Skills       []string        `json:"skills"`
}

type UpdateExtracurricularDTO struct {
	ActivityName *string          `json:"activity_name" validate:"omitempty,min=1,max=255"`
	Category     *models.Category `json:"category" validate:"omitempty,oneof=academic sports music arts community_service leadership technology volunteering clubs work_experience research other"`
	Position     *string          `json:"position" validate:"omitempty,min=1,max=255"`
	HoursPerWeek *int             `json:"hours_per_week" validate:"omitempty,min=1,max=168"`
	StartDate    *string          `json:"start_date" validate:"omitempty,len=7"` // YYYY-MM format
	EndDate      *string          `json:"end_date" validate:"omitempty"`         // YYYY-MM format or "present"
	Description  *string          `json:"description" validate:"omitempty,max=1000"`
	Location     *string          `json:"location" validate:"omitempty,max=255"`
	Website      *string          `json:"website" validate:"omitempty,url,max=255"`
	AIScore      *string          `json:"ai_score" validate:"omitempty,oneof=high medium low"`
	Achievements []string         `json:"achievements"`
	Skills       []string         `json:"skills"`
}

type ExtracurricularResponseDTO struct {
	ID           string          `json:"id"`
	ActivityName string          `json:"activity_name"`
	Category     models.Category `json:"category"`
	Position     string          `json:"position"`
	HoursPerWeek int             `json:"hours_per_week"`
	StartDate    string          `json:"start_date"`
	EndDate      string          `json:"end_date"`
	Description  string          `json:"description"`
	Location     string          `json:"location"`
	Website      string          `json:"website"`
	AIScore      string          `json:"ai_score"`
	Achievements []string        `json:"achievements"`
	Skills       []string        `json:"skills"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

// Conversion methods
func (dto *CreateExtracurricularDTO) ToModel(userID primitive.ObjectID) (*models.Extracurricular, error) {
	extracurricular := &models.Extracurricular{
		UserID:       userID.Hex(),
		ActivityName: dto.ActivityName,
		Category:     dto.Category,
		Position:     dto.Position,
		HoursPerWeek: dto.HoursPerWeek,
		StartDate:    dto.StartDate,
		EndDate:      dto.EndDate,
		Description:  dto.Description,
		Location:     dto.Location,
		Website:      dto.Website,
		Achievements: dto.Achievements,
		Skills:       dto.Skills,
	}

	extracurricular.BeforeCreate()
	return extracurricular, nil
}

func ExtracurricularToResponseDTO(extracurricular *models.Extracurricular) *ExtracurricularResponseDTO {
	return &ExtracurricularResponseDTO{
		ID:           extracurricular.ID.Hex(),
		ActivityName: extracurricular.ActivityName,
		Category:     extracurricular.Category,
		Position:     extracurricular.Position,
		HoursPerWeek: extracurricular.HoursPerWeek,
		StartDate:    extracurricular.StartDate,
		EndDate:      extracurricular.EndDate,
		Description:  extracurricular.Description,
		Location:     extracurricular.Location,
		Website:      extracurricular.Website,
		AIScore:      extracurricular.AIScore,
		Achievements: extracurricular.Achievements,
		Skills:       extracurricular.Skills,
		CreatedAt:    extracurricular.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    extracurricular.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToUpdateMap converts UpdateExtracurricularDTO to a map for database updates
func (dto *UpdateExtracurricularDTO) ToUpdateMap() (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if dto.ActivityName != nil {
		updates["activity_name"] = *dto.ActivityName
	}
	if dto.Category != nil {
		updates["category"] = *dto.Category
	}
	if dto.Position != nil {
		updates["position"] = *dto.Position
	}
	if dto.HoursPerWeek != nil {
		updates["hours_per_week"] = *dto.HoursPerWeek
	}
	if dto.StartDate != nil {
		updates["start_date"] = *dto.StartDate
	}
	if dto.EndDate != nil {
		updates["end_date"] = *dto.EndDate
	}
	if dto.Description != nil {
		updates["description"] = *dto.Description
	}
	if dto.Location != nil {
		updates["location"] = *dto.Location
	}
	if dto.Website != nil {
		updates["website"] = *dto.Website
	}
	if dto.AIScore != nil {
		updates["ai_score"] = *dto.AIScore
	}
	if dto.Achievements != nil {
		updates["achievements"] = dto.Achievements
	}
	if dto.Skills != nil {
		updates["skills"] = dto.Skills
	}

	return updates, nil
}
