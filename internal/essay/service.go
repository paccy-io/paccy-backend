package essay

import (
	"context"
	"reflect"

	"github.com/go-playground/validator/v10"
	collegeRepo "github.com/paccy-io/paccy-backend/internal/college"
	"github.com/paccy-io/paccy-backend/internal/essay/dto"
	"github.com/paccy-io/paccy-backend/internal/essay/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	essayServiceLogger = logger.New("essay")
	ErrNotFound        = errors.ErrGetFailed // using generic error for essays
)

type EssayService interface {
	Create(ctx context.Context, userID string, createDTO *dto.CreateEssayDTO) (*dto.EssayResponseDTO, error)
	GetByID(ctx context.Context, userID string, id string) (*dto.EssayResponseDTO, error)
	GetByUserID(ctx context.Context, userID string) ([]*dto.EssayResponseDTO, error)
	GetByCollegeApplicationID(ctx context.Context, userID string, collegeApplicationID string) ([]*dto.EssayResponseDTO, error)
	Update(ctx context.Context, userID string, id string, updateDTO *dto.UpdateEssayDTO) (*dto.EssayResponseDTO, error)
	Delete(ctx context.Context, userID string, id string) error
	GetByStatus(ctx context.Context, userID string, status string) ([]*dto.EssayResponseDTO, error)
	GetEssaysWithDeadlines(ctx context.Context, userID string, days int) ([]*dto.EssayResponseDTO, error)
	Count(ctx context.Context, userID string) (int64, error)
}

type essayService struct {
	repo        EssayRepository
	collegeRepo collegeRepo.CollegeRepository
	validator   *validator.Validate
}

func NewEssayService(repo EssayRepository, collegeRepo collegeRepo.CollegeRepository, validator *validator.Validate) EssayService {
	return &essayService{
		repo:        repo,
		collegeRepo: collegeRepo,
		validator:   validator,
	}
}

func (s *essayService) Create(ctx context.Context, userID string, createDTO *dto.CreateEssayDTO) (*dto.EssayResponseDTO, error) {
	essayServiceLogger.Info("Creating essay", map[string]interface{}{
		"user_id":                userID,
		"title":                  createDTO.Title,
		"college_application_id": createDTO.CollegeApplicationID,
	})

	// Validate DTO
	if err := s.validator.Struct(createDTO); err != nil {
		essayServiceLogger.Error("Essay validation failed", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify that the college application exists and user owns it
	_, err := s.collegeRepo.GetByID(ctx, createDTO.CollegeApplicationID)
	if err != nil {
		essayServiceLogger.Error("College application not found or access denied", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": createDTO.CollegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		essayServiceLogger.Error("Invalid user ID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Convert DTO to model
	essay, err := createDTO.ToModel(userObjectID)
	if err != nil {
		essayServiceLogger.Error("Failed to convert DTO to model", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Create the essay
	createdEssay, err := s.repo.Create(ctx, essay)
	if err != nil {
		essayServiceLogger.Error("Failed to create essay", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	// Get college name for response
	collegeApp, err := s.collegeRepo.GetByID(ctx, createDTO.CollegeApplicationID)
	if err != nil {
		essayServiceLogger.Warn("Failed to get college name for response", map[string]interface{}{
			"college_application_id": createDTO.CollegeApplicationID,
			"error":                  err.Error(),
		})
	}

	collegeName := ""
	if collegeApp != nil {
		collegeName = collegeApp.Name
	}

	essayServiceLogger.Info("Essay created successfully", map[string]interface{}{
		"user_id":  userID,
		"essay_id": createdEssay.ID.Hex(),
	})

	return dto.EssayToResponseDTO(createdEssay, collegeName), nil
}

func (s *essayService) GetByID(ctx context.Context, userID string, id string) (*dto.EssayResponseDTO, error) {
	essay, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership by checking if the college application belongs to the user
	collegeApp, err := s.collegeRepo.GetByID(ctx, essay.CollegeApplicationID.Hex())
	if err != nil {
		essayServiceLogger.Warn("College application not found for essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
			"error":    err.Error(),
		})
		return nil, ErrNotFound
	}

	if collegeApp.UserID.Hex() != userID {
		essayServiceLogger.Warn("Unauthorized access attempt to essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
			"owner_id": collegeApp.UserID.Hex(),
		})
		return nil, ErrNotFound // Don't reveal existence
	}

	return dto.EssayToResponseDTO(essay, collegeApp.Name), nil
}

func (s *essayService) GetByUserID(ctx context.Context, userID string) ([]*dto.EssayResponseDTO, error) {
	essays, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.convertEssaysToResponseDTOs(ctx, essays)
}

func (s *essayService) GetByCollegeApplicationID(ctx context.Context, userID string, collegeApplicationID string) ([]*dto.EssayResponseDTO, error) {
	// First verify that the user owns the college application
	collegeApp, err := s.collegeRepo.GetByID(ctx, collegeApplicationID)
	if err != nil {
		essayServiceLogger.Error("College application not found", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, err
	}

	if collegeApp.UserID.Hex() != userID {
		essayServiceLogger.Warn("Unauthorized access attempt to college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"owner_id":               collegeApp.UserID.Hex(),
		})
		return nil, ErrNotFound
	}

	essays, err := s.repo.GetByCollegeApplicationID(ctx, collegeApplicationID)
	if err != nil {
		return nil, err
	}

	// Convert essays to response DTOs with college name
	responseDTOs := make([]*dto.EssayResponseDTO, len(essays))
	for i, essay := range essays {
		responseDTOs[i] = dto.EssayToResponseDTO(essay, collegeApp.Name)
	}

	return responseDTOs, nil
}

func (s *essayService) Update(ctx context.Context, userID string, id string, updateDTO *dto.UpdateEssayDTO) (*dto.EssayResponseDTO, error) {
	essayServiceLogger.Info("Updating essay", map[string]interface{}{
		"user_id":  userID,
		"essay_id": id,
	})

	// Validate DTO
	if err := s.validator.Struct(updateDTO); err != nil {
		essayServiceLogger.Error("Essay update validation failed", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
			"error":    err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// First, verify the essay exists and user owns it
	existing, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	// Build update map from DTO
	updates := bson.M{}

	// Use reflection to check for non-nil pointers and set them
	v := reflect.ValueOf(updateDTO).Elem()
	t := reflect.TypeOf(updateDTO).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the json tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Handle pointer fields
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			updates[jsonTag] = field.Elem().Interface()
		} else if field.Kind() == reflect.Slice && !field.IsNil() && field.Len() > 0 {
			// Handle slices (for AI feedbacks)
			updates[jsonTag] = field.Interface()
		}
	}

	if len(updates) == 0 {
		essayServiceLogger.Warn("No fields to update", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
		})
		return existing, nil
	}

	// Update the essay
	updatedEssay, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		essayServiceLogger.Error("Failed to update essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Get college name for response
	collegeApp, err := s.collegeRepo.GetByID(ctx, updatedEssay.CollegeApplicationID.Hex())
	collegeName := ""
	if err == nil && collegeApp != nil {
		collegeName = collegeApp.Name
	}

	essayServiceLogger.Info("Essay updated successfully", map[string]interface{}{
		"user_id":  userID,
		"essay_id": id,
	})

	return dto.EssayToResponseDTO(updatedEssay, collegeName), nil
}

func (s *essayService) Delete(ctx context.Context, userID string, id string) error {
	essayServiceLogger.Info("Deleting essay", map[string]interface{}{
		"user_id":  userID,
		"essay_id": id,
	})

	// First, verify the essay exists and user owns it
	_, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Delete the essay
	err = s.repo.Delete(ctx, id)
	if err != nil {
		essayServiceLogger.Error("Failed to delete essay", map[string]interface{}{
			"user_id":  userID,
			"essay_id": id,
			"error":    err.Error(),
		})
		return err
	}

	essayServiceLogger.Info("Essay deleted successfully", map[string]interface{}{
		"user_id":  userID,
		"essay_id": id,
	})

	return nil
}

func (s *essayService) GetByStatus(ctx context.Context, userID string, status string) ([]*dto.EssayResponseDTO, error) {
	// Validate status
	if models.IsValidEssayStatus(status) == "" {
		return nil, errors.ErrInvalidInput
	}

	essays, err := s.repo.GetByStatus(ctx, userID, status)
	if err != nil {
		return nil, err
	}

	return s.convertEssaysToResponseDTOs(ctx, essays)
}

func (s *essayService) GetEssaysWithDeadlines(ctx context.Context, userID string, days int) ([]*dto.EssayResponseDTO, error) {
	essays, err := s.repo.GetEssaysWithDeadlines(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	return s.convertEssaysToResponseDTOs(ctx, essays)
}

func (s *essayService) Count(ctx context.Context, userID string) (int64, error) {
	return s.repo.Count(ctx, userID)
}

// Helper method to convert essays to response DTOs with college names
func (s *essayService) convertEssaysToResponseDTOs(ctx context.Context, essays []*models.Essay) ([]*dto.EssayResponseDTO, error) {
	responseDTOs := make([]*dto.EssayResponseDTO, len(essays))

	for i, essay := range essays {
		// Get college name for each essay
		collegeApp, err := s.collegeRepo.GetByID(ctx, essay.CollegeApplicationID.Hex())
		collegeName := ""
		if err == nil && collegeApp != nil {
			collegeName = collegeApp.Name
		} else {
			essayServiceLogger.Warn("Failed to get college name for essay", map[string]interface{}{
				"essay_id":               essay.ID.Hex(),
				"college_application_id": essay.CollegeApplicationID.Hex(),
				"error":                  err.Error(),
			})
		}

		responseDTOs[i] = dto.EssayToResponseDTO(essay, collegeName)
	}

	return responseDTOs, nil
}
