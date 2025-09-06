package dto

import (
	"github.com/paccy-io/paccy-backend/internal/academic/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Exam DTOs
type CreateExamDTO struct {
	Type     models.ExamType `json:"type" validate:"required,oneof=SAT ACT 'SAT Subject Test' AP IB TOEFL IELTS GRE GMAT LSAT MCAT other"`
	Name     string          `json:"name" validate:"required,min=1,max=255"`
	Score    string          `json:"score" validate:"required,min=1,max=50"`
	MaxScore string          `json:"max_score" validate:"required,min=1,max=50"`
	ExamDate string          `json:"exam_date" validate:"required,len=7"` // YYYY-MM format
	Notes    string          `json:"notes" validate:"max=1000"`
}

type UpdateExamDTO struct {
	Type     *models.ExamType `json:"type" validate:"omitempty,oneof=SAT ACT 'SAT Subject Test' AP IB TOEFL IELTS GRE GMAT LSAT MCAT other"`
	Name     *string          `json:"name" validate:"omitempty,min=1,max=255"`
	Score    *string          `json:"score" validate:"omitempty,min=1,max=50"`
	MaxScore *string          `json:"max_score" validate:"omitempty,min=1,max=50"`
	ExamDate *string          `json:"exam_date" validate:"omitempty,len=7"` // YYYY-MM format
	Notes    *string          `json:"notes" validate:"omitempty,max=1000"`
}

type ExamResponseDTO struct {
	ID        string          `json:"id"`
	Type      models.ExamType `json:"type"`
	Name      string          `json:"name"`
	Score     string          `json:"score"`
	MaxScore  string          `json:"max_score"`
	ExamDate  string          `json:"exam_date"`
	Notes     string          `json:"notes"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// HighSchoolGPA DTOs
type CreateHighSchoolGPADTO struct {
	Scale models.GPAScale `json:"scale" validate:"omitempty,oneof=4.0 5.0 100"`
	GPA1  string          `json:"gpa_1" validate:"omitempty,max=10"`
	GPA2  string          `json:"gpa_2" validate:"omitempty,max=10"`
	GPA3  string          `json:"gpa_3" validate:"omitempty,max=10"`
	GPA4  string          `json:"gpa_4" validate:"omitempty,max=10"`
}

type UpdateHighSchoolGPADTO struct {
	Scale *models.GPAScale `json:"scale" validate:"omitempty,oneof=4.0 5.0 100"`
	GPA1  *string          `json:"gpa_1" validate:"omitempty,max=10"`
	GPA2  *string          `json:"gpa_2" validate:"omitempty,max=10"`
	GPA3  *string          `json:"gpa_3" validate:"omitempty,max=10"`
	GPA4  *string          `json:"gpa_4" validate:"omitempty,max=10"`
}

type HighSchoolGPAResponseDTO struct {
	ID        string          `json:"id"`
	Scale     models.GPAScale `json:"scale"`
	GPA1      string          `json:"gpa_1"`
	GPA2      string          `json:"gpa_2"`
	GPA3      string          `json:"gpa_3"`
	GPA4      string          `json:"gpa_4"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

// Exam conversion methods
func (dto *CreateExamDTO) ToModel(userID primitive.ObjectID) (*models.Exam, error) {
	exam := &models.Exam{
		UserID:   userID.Hex(),
		Type:     dto.Type,
		Name:     dto.Name,
		Score:    dto.Score,
		MaxScore: dto.MaxScore,
		ExamDate: dto.ExamDate,
		Notes:    dto.Notes,
	}

	exam.BeforeCreate()
	return exam, nil
}

func ExamToResponseDTO(exam *models.Exam) *ExamResponseDTO {
	return &ExamResponseDTO{
		ID:        exam.ID.Hex(),
		Type:      exam.Type,
		Name:      exam.Name,
		Score:     exam.Score,
		MaxScore:  exam.MaxScore,
		ExamDate:  exam.ExamDate,
		Notes:     exam.Notes,
		CreatedAt: exam.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: exam.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (dto *UpdateExamDTO) ToUpdateMap() (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if dto.Type != nil {
		updates["type"] = *dto.Type
	}
	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.Score != nil {
		updates["score"] = *dto.Score
	}
	if dto.MaxScore != nil {
		updates["max_score"] = *dto.MaxScore
	}
	if dto.ExamDate != nil {
		updates["exam_date"] = *dto.ExamDate
	}
	if dto.Notes != nil {
		updates["notes"] = *dto.Notes
	}

	return updates, nil
}

// HighSchoolGPA conversion methods
func (dto *CreateHighSchoolGPADTO) ToModel(userID primitive.ObjectID) (*models.HighSchoolGPA, error) {
	gpa := &models.HighSchoolGPA{
		UserID: userID.Hex(),
		Scale:  dto.Scale,
		GPA1:   dto.GPA1,
		GPA2:   dto.GPA2,
		GPA3:   dto.GPA3,
		GPA4:   dto.GPA4,
	}

	gpa.BeforeCreate()
	return gpa, nil
}

func HighSchoolGPAToResponseDTO(gpa *models.HighSchoolGPA) *HighSchoolGPAResponseDTO {
	return &HighSchoolGPAResponseDTO{
		ID:        gpa.ID.Hex(),
		Scale:     gpa.Scale,
		GPA1:      gpa.GPA1,
		GPA2:      gpa.GPA2,
		GPA3:      gpa.GPA3,
		GPA4:      gpa.GPA4,
		CreatedAt: gpa.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: gpa.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (dto *UpdateHighSchoolGPADTO) ToUpdateMap() (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if dto.Scale != nil {
		updates["scale"] = *dto.Scale
	}
	if dto.GPA1 != nil {
		updates["gpa_1"] = *dto.GPA1
	}
	if dto.GPA2 != nil {
		updates["gpa_2"] = *dto.GPA2
	}
	if dto.GPA3 != nil {
		updates["gpa_3"] = *dto.GPA3
	}
	if dto.GPA4 != nil {
		updates["gpa_4"] = *dto.GPA4
	}

	return updates, nil
}
