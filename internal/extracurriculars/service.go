package extracurriculars

import (
	"context"
	"time"

	"github.com/paccy-io/paccy-backend/internal/extracurriculars/dto"
	"github.com/paccy-io/paccy-backend/internal/extracurriculars/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var serviceLogger = logger.New("extracurricular_service")

type ExtracurricularService interface {
	CreateExtracurricular(ctx context.Context, userID primitive.ObjectID, req *dto.CreateExtracurricularDTO) (*dto.ExtracurricularResponseDTO, error)
	GetExtracurricularByID(ctx context.Context, id string, userID string) (*dto.ExtracurricularResponseDTO, error)
	GetExtracurricularsByUserID(ctx context.Context, userID string) ([]*dto.ExtracurricularResponseDTO, error)
	GetExtracurricularsByCategory(ctx context.Context, userID string, category string) ([]*dto.ExtracurricularResponseDTO, error)
	UpdateExtracurricular(ctx context.Context, id string, userID string, req *dto.UpdateExtracurricularDTO) (*dto.ExtracurricularResponseDTO, error)
	DeleteExtracurricular(ctx context.Context, id string, userID string) error
}

type extracurricularService struct {
	repo ExtracurricularRepository
}

func NewExtracurricularService(repo ExtracurricularRepository) ExtracurricularService {
	return &extracurricularService{
		repo: repo,
	}
}

func (s *extracurricularService) CreateExtracurricular(ctx context.Context, userID primitive.ObjectID, req *dto.CreateExtracurricularDTO) (*dto.ExtracurricularResponseDTO, error) {
	serviceLogger.Info("Creating new extracurricular", map[string]interface{}{
		"activity_name": req.ActivityName,
		"user_id":       userID.Hex(),
		"category":      req.Category,
	})

	// Convert DTO to model
	extracurricular, err := req.ToModel(userID)
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to model", map[string]interface{}{
			"error":         err.Error(),
			"activity_name": req.ActivityName,
			"user_id":       userID.Hex(),
		})
		return nil, err
	}

	// Create extracurricular
	createdExtracurricular, err := s.repo.Create(ctx, extracurricular)
	if err != nil {
		serviceLogger.Error("Failed to create extracurricular in repository", map[string]interface{}{
			"error":         err.Error(),
			"activity_name": req.ActivityName,
			"user_id":       userID.Hex(),
		})
		return nil, err
	}

	serviceLogger.Info("Extracurricular created successfully", map[string]interface{}{
		"extracurricular_id": createdExtracurricular.ID.Hex(),
		"activity_name":      createdExtracurricular.ActivityName,
		"user_id":            createdExtracurricular.UserID,
	})

	return dto.ExtracurricularToResponseDTO(createdExtracurricular), nil
}

func (s *extracurricularService) GetExtracurricularByID(ctx context.Context, id string, userID string) (*dto.ExtracurricularResponseDTO, error) {
	serviceLogger.Info("Getting extracurricular by ID", map[string]interface{}{
		"extracurricular_id": id,
		"user_id":            userID,
	})

	extracurricular, err := s.repo.GetByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get extracurricular by ID", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return nil, err
	}

	// Check ownership
	if extracurricular.UserID != userID {
		serviceLogger.Warn("User attempted to access extracurricular they don't own", map[string]interface{}{
			"extracurricular_id":    id,
			"requesting_user_id":    userID,
			"extracurricular_owner": extracurricular.UserID,
		})
		return nil, errors.ErrGetFailed
	}

	return dto.ExtracurricularToResponseDTO(extracurricular), nil
}

func (s *extracurricularService) GetExtracurricularsByUserID(ctx context.Context, userID string) ([]*dto.ExtracurricularResponseDTO, error) {
	serviceLogger.Info("Getting extracurriculars by user ID", map[string]interface{}{
		"user_id": userID,
	})

	extracurriculars, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		serviceLogger.Error("Failed to get extracurriculars by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, err
	}

	var responseDTOs []*dto.ExtracurricularResponseDTO
	for _, extracurricular := range extracurriculars {
		responseDTOs = append(responseDTOs, dto.ExtracurricularToResponseDTO(extracurricular))
	}

	serviceLogger.Info("Retrieved extracurriculars for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(responseDTOs),
	})

	return responseDTOs, nil
}

func (s *extracurricularService) GetExtracurricularsByCategory(ctx context.Context, userID string, category string) ([]*dto.ExtracurricularResponseDTO, error) {
	serviceLogger.Info("Getting extracurriculars by category", map[string]interface{}{
		"user_id":  userID,
		"category": category,
	})

	// Validate category
	if models.IsValidCategory(category) == "" {
		serviceLogger.Error("Invalid extracurricular category", map[string]interface{}{
			"category": category,
			"user_id":  userID,
		})
		return nil, errors.ErrInvalidInput
	}

	extracurriculars, err := s.repo.GetByCategory(ctx, userID, category)
	if err != nil {
		serviceLogger.Error("Failed to get extracurriculars by category", map[string]interface{}{
			"error":    err.Error(),
			"user_id":  userID,
			"category": category,
		})
		return nil, err
	}

	var responseDTOs []*dto.ExtracurricularResponseDTO
	for _, extracurricular := range extracurriculars {
		responseDTOs = append(responseDTOs, dto.ExtracurricularToResponseDTO(extracurricular))
	}

	serviceLogger.Info("Retrieved extracurriculars by category", map[string]interface{}{
		"user_id":  userID,
		"category": category,
		"count":    len(responseDTOs),
	})

	return responseDTOs, nil
}

func (s *extracurricularService) UpdateExtracurricular(ctx context.Context, id string, userID string, req *dto.UpdateExtracurricularDTO) (*dto.ExtracurricularResponseDTO, error) {
	serviceLogger.Info("Updating extracurricular", map[string]interface{}{
		"extracurricular_id": id,
		"user_id":            userID,
	})

	// Check if extracurricular exists and user owns it
	extracurricular, err := s.repo.GetByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get extracurricular for update", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return nil, err
	}

	// Check ownership
	if extracurricular.UserID != userID {
		serviceLogger.Warn("User attempted to update extracurricular they don't own", map[string]interface{}{
			"extracurricular_id":    id,
			"requesting_user_id":    userID,
			"extracurricular_owner": extracurricular.UserID,
		})
		return nil, errors.ErrGetFailed
	}

	// Convert DTO to update map
	updateMap, err := req.ToUpdateMap()
	if err != nil {
		serviceLogger.Error("Failed to convert DTO to update map", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return nil, err
	}
	updateMap["updated_at"] = time.Now()

	// Update extracurricular
	updatedExtracurricular, err := s.repo.Update(ctx, id, updateMap)
	if err != nil {
		serviceLogger.Error("Failed to update extracurricular in repository", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return nil, err
	}

	serviceLogger.Info("Extracurricular updated successfully", map[string]interface{}{
		"extracurricular_id": id,
		"activity_name":      updatedExtracurricular.ActivityName,
		"user_id":            userID,
	})

	return dto.ExtracurricularToResponseDTO(updatedExtracurricular), nil
}

func (s *extracurricularService) DeleteExtracurricular(ctx context.Context, id string, userID string) error {
	serviceLogger.Info("Deleting extracurricular", map[string]interface{}{
		"extracurricular_id": id,
		"user_id":            userID,
	})

	// Check if extracurricular exists and user owns it
	extracurricular, err := s.repo.GetByID(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to get extracurricular for deletion", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return err
	}

	// Check ownership
	if extracurricular.UserID != userID {
		serviceLogger.Warn("User attempted to delete extracurricular they don't own", map[string]interface{}{
			"extracurricular_id":    id,
			"requesting_user_id":    userID,
			"extracurricular_owner": extracurricular.UserID,
		})
		return errors.ErrGetFailed
	}

	// Delete extracurricular
	err = s.repo.Delete(ctx, id)
	if err != nil {
		serviceLogger.Error("Failed to delete extracurricular in repository", map[string]interface{}{
			"error":              err.Error(),
			"extracurricular_id": id,
			"user_id":            userID,
		})
		return err
	}

	serviceLogger.Info("Extracurricular deleted successfully", map[string]interface{}{
		"extracurricular_id": id,
		"user_id":            userID,
	})

	return nil
}
