package academic

import (
	"context"

	"github.com/paccy-io/paccy-backend/internal/academic/dto"
	"github.com/paccy-io/paccy-backend/internal/academic/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var serviceLogger = logger.New("academic_service")

type AcademicService interface {
	// Exam methods
	CreateExam(ctx context.Context, userID primitive.ObjectID, req *dto.CreateExamDTO) (*dto.ExamResponseDTO, error)
	GetExamByID(ctx context.Context, id string, userID string) (*dto.ExamResponseDTO, error)
	GetExamsByUserID(ctx context.Context, userID string) ([]*dto.ExamResponseDTO, error)
	GetExamsByType(ctx context.Context, userID string, examType string) ([]*dto.ExamResponseDTO, error)
	UpdateExam(ctx context.Context, id string, userID string, req *dto.UpdateExamDTO) (*dto.ExamResponseDTO, error)
	DeleteExam(ctx context.Context, id string, userID string) error

	// HighSchoolGPA methods
	CreateHighSchoolGPA(ctx context.Context, userID primitive.ObjectID, req *dto.CreateHighSchoolGPADTO) (*dto.HighSchoolGPAResponseDTO, error)
	GetHighSchoolGPAByUserID(ctx context.Context, userID string) (*dto.HighSchoolGPAResponseDTO, error)
	UpdateHighSchoolGPA(ctx context.Context, userID string, req *dto.UpdateHighSchoolGPADTO) (*dto.HighSchoolGPAResponseDTO, error)
	DeleteHighSchoolGPA(ctx context.Context, userID string) error
}

type academicService struct {
	repo AcademicRepository
}

func NewAcademicService(repo AcademicRepository) AcademicService {
	return &academicService{
		repo: repo,
	}
}

// Exam methods
func (s *academicService) CreateExam(ctx context.Context, userID primitive.ObjectID, req *dto.CreateExamDTO) (*dto.ExamResponseDTO, error) {
	serviceLogger.Info("Creating new exam", map[string]interface{}{
		"exam_name": req.Name,
		"exam_type": req.Type,
		"user_id":   userID.Hex(),
	})

	// Convert DTO to model
	exam, err := req.ToModel(userID)
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to model", map[string]interface{}{
			"error":     err.Error(),
			"exam_name": req.Name,
			"user_id":   userID.Hex(),
		})
		return nil, err
	}

	// Create exam in repository
	createdExam, err := s.repo.CreateExam(ctx, exam)
	if err != nil {
		serviceLogger.Error("Failed to create exam", map[string]interface{}{
			"error":     err.Error(),
			"exam_name": req.Name,
			"user_id":   userID.Hex(),
		})
		return nil, err
	}

	serviceLogger.Info("Exam created successfully", map[string]interface{}{
		"exam_id":   createdExam.ID.Hex(),
		"exam_name": createdExam.Name,
		"user_id":   userID.Hex(),
	})

	return dto.ExamToResponseDTO(createdExam), nil
}

func (s *academicService) GetExamByID(ctx context.Context, id string, userID string) (*dto.ExamResponseDTO, error) {
	serviceLogger.Info("Getting exam by ID", map[string]interface{}{
		"exam_id": id,
		"user_id": userID,
	})

	exam, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get exam by ID", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return nil, err
	}

	// Check ownership
	if exam.UserID != userID {
		serviceLogger.Warning("User attempted to access exam they don't own", map[string]interface{}{
			"exam_id":            id,
			"exam_owner":         exam.UserID,
			"requesting_user_id": userID,
		})
		return nil, errors.ErrGetFailed
	}

	serviceLogger.Info("Retrieved exam by ID", map[string]interface{}{
		"exam_id":   id,
		"exam_name": exam.Name,
		"user_id":   userID,
	})

	return dto.ExamToResponseDTO(exam), nil
}

func (s *academicService) GetExamsByUserID(ctx context.Context, userID string) ([]*dto.ExamResponseDTO, error) {
	serviceLogger.Info("Getting exams by user ID", map[string]interface{}{
		"user_id": userID,
	})

	exams, err := s.repo.GetExamsByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get exams by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	var responseExams []*dto.ExamResponseDTO
	for _, exam := range exams {
		responseExams = append(responseExams, dto.ExamToResponseDTO(exam))
	}

	serviceLogger.Info("Retrieved exams for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(exams),
	})

	return responseExams, nil
}

func (s *academicService) GetExamsByType(ctx context.Context, userID string, examType string) ([]*dto.ExamResponseDTO, error) {
	serviceLogger.Info("Getting exams by type", map[string]interface{}{
		"user_id":   userID,
		"exam_type": examType,
	})

	// Validate exam type
	if models.IsValidExamType(examType) == "" {
		serviceLogger.Error("Invalid exam type", map[string]interface{}{
			"exam_type": examType,
			"user_id":   userID,
		})
		return nil, errors.ErrInvalidInput
	}

	exams, err := s.repo.GetExamsByType(ctx, userID, examType)
	if err != nil {
		serviceLogger.Error("Failed to get exams by type", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID,
			"exam_type": examType,
		})
		return nil, err
	}

	var responseExams []*dto.ExamResponseDTO
	for _, exam := range exams {
		responseExams = append(responseExams, dto.ExamToResponseDTO(exam))
	}

	serviceLogger.Info("Retrieved exams by type for user", map[string]interface{}{
		"user_id":   userID,
		"exam_type": examType,
		"count":     len(exams),
	})

	return responseExams, nil
}

func (s *academicService) UpdateExam(ctx context.Context, id string, userID string, req *dto.UpdateExamDTO) (*dto.ExamResponseDTO, error) {
	serviceLogger.Info("Updating exam", map[string]interface{}{
		"exam_id": id,
		"user_id": userID,
	})

	// First, get the exam to check ownership
	exam, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get exam for update", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return nil, err
	}

	// Check ownership
	if exam.UserID != userID {
		serviceLogger.Warning("User attempted to update exam they don't own", map[string]interface{}{
			"exam_id":            id,
			"exam_owner":         exam.UserID,
			"requesting_user_id": userID,
		})
		return nil, errors.ErrUpdateFailed
	}

	// Convert DTO to update map
	updates, err := req.ToUpdateMap()
	if err != nil {
		serviceLogger.Error("Failed to convert update DTO", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return nil, err
	}

	// Update exam in repository
	updatedExam, err := s.repo.UpdateExam(ctx, id, updates)
	if err != nil {
		serviceLogger.Error("Failed to update exam", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return nil, err
	}

	serviceLogger.Info("Exam updated successfully", map[string]interface{}{
		"exam_id":   id,
		"exam_name": updatedExam.Name,
		"user_id":   userID,
	})

	return dto.ExamToResponseDTO(updatedExam), nil
}

func (s *academicService) DeleteExam(ctx context.Context, id string, userID string) error {
	serviceLogger.Info("Deleting exam", map[string]interface{}{
		"exam_id": id,
		"user_id": userID,
	})

	// First, get the exam to check ownership
	exam, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get exam for deletion", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return err
	}

	// Check ownership
	if exam.UserID != userID {
		serviceLogger.Warning("User attempted to delete exam they don't own", map[string]interface{}{
			"exam_id":            id,
			"exam_owner":         exam.UserID,
			"requesting_user_id": userID,
		})
		return errors.ErrDeleteFailed
	}

	// Delete exam in repository
	err = s.repo.DeleteExam(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to delete exam", map[string]interface{}{
			"error":   err.Error(),
			"exam_id": id,
			"user_id": userID,
		})
		return err
	}

	serviceLogger.Info("Exam deleted successfully", map[string]interface{}{
		"exam_id": id,
		"user_id": userID,
	})

	return nil
}

// HighSchoolGPA methods
func (s *academicService) CreateHighSchoolGPA(ctx context.Context, userID primitive.ObjectID, req *dto.CreateHighSchoolGPADTO) (*dto.HighSchoolGPAResponseDTO, error) {
	serviceLogger.Info("Creating new HighSchoolGPA", map[string]interface{}{
		"user_id": userID.Hex(),
		"scale":   req.Scale,
	})

	// Validate scale if provided
	if req.Scale != "" {
		if validScale := models.IsValidGPAScale(string(req.Scale)); validScale == "" {
			serviceLogger.Error("Invalid GPA scale provided", map[string]interface{}{
				"scale":   req.Scale,
				"user_id": userID.Hex(),
			})
			return nil, errors.ErrInvalidInput
		}
	}

	// Check if user already has a GPA record (max 1 per user)
	existingGPA, err := s.repo.GetHighSchoolGPAByUserID(ctx, userID.Hex())
	if err == nil && existingGPA != nil {
		serviceLogger.Error("User already has a HighSchoolGPA record", map[string]interface{}{
			"user_id":         userID.Hex(),
			"existing_gpa_id": existingGPA.ID.Hex(),
		})
		return nil, errors.ErrUserExists // Using this as "GPA already exists"
	}

	// Convert DTO to model
	gpa, err := req.ToModel(userID)
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to model", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.Hex(),
		})
		return nil, err
	}

	// Create GPA in repository
	createdGPA, err := s.repo.CreateHighSchoolGPA(ctx, gpa)
	if err != nil {
		serviceLogger.Error("Failed to create HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.Hex(),
		})
		return nil, err
	}

	serviceLogger.Info("HighSchoolGPA created successfully", map[string]interface{}{
		"gpa_id":  createdGPA.ID.Hex(),
		"user_id": userID.Hex(),
	})

	return dto.HighSchoolGPAToResponseDTO(createdGPA), nil
}

func (s *academicService) GetHighSchoolGPAByUserID(ctx context.Context, userID string) (*dto.HighSchoolGPAResponseDTO, error) {
	serviceLogger.Info("Getting HighSchoolGPA by user ID", map[string]interface{}{
		"user_id": userID,
	})

	gpa, err := s.repo.GetHighSchoolGPAByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get HighSchoolGPA by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	serviceLogger.Info("Retrieved HighSchoolGPA for user", map[string]interface{}{
		"user_id": userID,
		"gpa_id":  gpa.ID.Hex(),
	})

	return dto.HighSchoolGPAToResponseDTO(gpa), nil
}

func (s *academicService) UpdateHighSchoolGPA(ctx context.Context, userID string, req *dto.UpdateHighSchoolGPADTO) (*dto.HighSchoolGPAResponseDTO, error) {
	serviceLogger.Info("Updating HighSchoolGPA", map[string]interface{}{
		"user_id": userID,
	})

	// Validate scale if provided
	if req.Scale != nil && *req.Scale != "" {
		if validScale := models.IsValidGPAScale(string(*req.Scale)); validScale == "" {
			serviceLogger.Error("Invalid GPA scale provided for update", map[string]interface{}{
				"scale":   *req.Scale,
				"user_id": userID,
			})
			return nil, errors.ErrInvalidInput
		}
	}

	// Get the existing GPA record
	gpa, err := s.repo.GetHighSchoolGPAByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get HighSchoolGPA for update", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	// Convert DTO to update map
	updates, err := req.ToUpdateMap()
	if err != nil {
		serviceLogger.Error("Failed to convert update DTO", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	// Update GPA in repository
	updatedGPA, err := s.repo.UpdateHighSchoolGPA(ctx, gpa.ID.Hex(), updates)
	if err != nil {
		serviceLogger.Error("Failed to update HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	serviceLogger.Info("HighSchoolGPA updated successfully", map[string]interface{}{
		"gpa_id":  gpa.ID.Hex(),
		"user_id": userID,
	})

	return dto.HighSchoolGPAToResponseDTO(updatedGPA), nil
}

func (s *academicService) DeleteHighSchoolGPA(ctx context.Context, userID string) error {
	serviceLogger.Info("Deleting HighSchoolGPA", map[string]interface{}{
		"user_id": userID,
	})

	// Get the existing GPA record
	gpa, err := s.repo.GetHighSchoolGPAByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get HighSchoolGPA for deletion", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return err
	}

	// Delete GPA in repository
	err = s.repo.DeleteHighSchoolGPA(ctx, gpa.ID.Hex())
	if err != nil {
		serviceLogger.Error("Failed to delete HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return err
	}

	serviceLogger.Info("HighSchoolGPA deleted successfully", map[string]interface{}{
		"gpa_id":  gpa.ID.Hex(),
		"user_id": userID,
	})

	return nil
}
