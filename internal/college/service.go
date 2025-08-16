package college

import (
	"context"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/college/dto"
	"github.com/paccy-io/paccy-backend/internal/college/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	collegeServiceLogger = logger.New("college")
	ErrNotFound          = errors.ErrCollegeNotFound // using specific college error
)

type CollegeService interface {
	Create(ctx context.Context, userID string, createDTO *dto.CreateCollegeApplicationDTO) (*models.CollegeApplication, error)
	GetByID(ctx context.Context, userID string, id string) (*models.CollegeApplication, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.CollegeApplication, error)
	Update(ctx context.Context, userID string, id string, updateDTO *dto.UpdateCollegeApplicationDTO) (*models.CollegeApplication, error)
	Delete(ctx context.Context, userID string, id string) error
	GetByStatus(ctx context.Context, userID string, status string) ([]*models.CollegeApplication, error)
	GetByPriority(ctx context.Context, userID string, priority string) ([]*models.CollegeApplication, error)
	GetUpcomingDeadlines(ctx context.Context, userID string, days int) ([]*models.CollegeApplication, error)
	Count(ctx context.Context, userID string) (int64, error)

	// Checklist operations
	CreateChecklist(ctx context.Context, userID string, createDTO *dto.StandaloneCreateChecklistDTO) (*models.Checklist, error)
	GetChecklist(ctx context.Context, userID string, applicationID string, checklistID string) (*models.Checklist, error)
	UpdateChecklist(ctx context.Context, userID string, applicationID string, checklistID string, updateDTO *dto.StandaloneUpdateChecklistDTO) (*models.Checklist, error)
	DeleteChecklist(ctx context.Context, userID string, applicationID string, checklistID string) error

	// Task operations
	CreateTask(ctx context.Context, userID string, createDTO *dto.StandaloneCreateTaskDTO) (*models.Task, error)
	GetTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string) (*models.Task, error)
	UpdateTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string, updateDTO *dto.StandaloneUpdateTaskDTO) (*models.Task, error)
	DeleteTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string) error
}

type collegeService struct {
	repo      CollegeRepository
	validator *validator.Validate
}

func NewCollegeService(repo CollegeRepository, validator *validator.Validate) CollegeService {
	return &collegeService{
		repo:      repo,
		validator: validator,
	}
}

func (s *collegeService) Create(ctx context.Context, userID string, createDTO *dto.CreateCollegeApplicationDTO) (*models.CollegeApplication, error) {
	collegeServiceLogger.Info("Creating college application", map[string]interface{}{
		"user_id": userID,
		"name":    createDTO.Name,
	})

	// Validate DTO
	if err := s.validator.Struct(createDTO); err != nil {
		collegeServiceLogger.Error("College application validation failed", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Convert userID to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeServiceLogger.Error("Invalid user ID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Convert DTO to model
	application := createDTO.ToModel(userObjectID)

	// Create the application
	createdApplication, err := s.repo.Create(ctx, application)
	if err != nil {
		collegeServiceLogger.Error("Failed to create college application", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	collegeServiceLogger.Info("College application created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": createdApplication.ID.Hex(),
	})

	return createdApplication, nil
}

func (s *collegeService) GetByID(ctx context.Context, userID string, id string) (*models.CollegeApplication, error) {
	application, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if application.UserID.Hex() != userID {
		collegeServiceLogger.Warn("Unauthorized access attempt to college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": id,
			"owner_id":       application.UserID.Hex(),
		})
		return nil, ErrNotFound // Don't reveal existence
	}

	return application, nil
}

func (s *collegeService) GetByUserID(ctx context.Context, userID string) ([]*models.CollegeApplication, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *collegeService) Update(ctx context.Context, userID string, id string, updateDTO *dto.UpdateCollegeApplicationDTO) (*models.CollegeApplication, error) {
	collegeServiceLogger.Info("Updating college application", map[string]interface{}{
		"user_id":        userID,
		"application_id": id,
	})

	// Validate DTO
	if err := s.validator.Struct(updateDTO); err != nil {
		collegeServiceLogger.Error("College application update validation failed", map[string]interface{}{
			"user_id":        userID,
			"application_id": id,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// First, verify the application exists and user owns it
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

		// Get the bson tag
		bsonTag := fieldType.Tag.Get("json")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		// Handle pointer fields
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			updates[bsonTag] = field.Elem().Interface()
		} else if field.Kind() == reflect.Slice && !field.IsNil() && field.Len() > 0 {
			// Handle slices (for arrays like Websites, Tags, Checklists)
			updates[bsonTag] = field.Interface()
		}
	}

	if len(updates) == 0 {
		collegeServiceLogger.Warn("No fields to update", map[string]interface{}{
			"user_id":        userID,
			"application_id": id,
		})
		return existing, nil
	}

	// Update the application
	updatedApplication, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		collegeServiceLogger.Error("Failed to update college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": id,
			"error":          err.Error(),
		})
		return nil, err
	}

	collegeServiceLogger.Info("College application updated successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": id,
	})

	return updatedApplication, nil
}

func (s *collegeService) Delete(ctx context.Context, userID string, id string) error {
	collegeServiceLogger.Info("Deleting college application", map[string]interface{}{
		"user_id":        userID,
		"application_id": id,
	})

	// First, verify the application exists and user owns it
	_, err := s.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Delete the application
	err = s.repo.Delete(ctx, id)
	if err != nil {
		collegeServiceLogger.Error("Failed to delete college application", map[string]interface{}{
			"user_id":        userID,
			"application_id": id,
			"error":          err.Error(),
		})
		return err
	}

	collegeServiceLogger.Info("College application deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": id,
	})

	return nil
}

func (s *collegeService) GetByStatus(ctx context.Context, userID string, status string) ([]*models.CollegeApplication, error) {
	return s.repo.GetByStatus(ctx, userID, status)
}

func (s *collegeService) GetByPriority(ctx context.Context, userID string, priority string) ([]*models.CollegeApplication, error) {
	return s.repo.GetByPriority(ctx, userID, priority)
}

func (s *collegeService) GetUpcomingDeadlines(ctx context.Context, userID string, days int) ([]*models.CollegeApplication, error) {
	return s.repo.GetUpcomingDeadlines(ctx, userID, days)
}

func (s *collegeService) Count(ctx context.Context, userID string) (int64, error) {
	return s.repo.Count(ctx, userID)
}

// Checklist operations
func (s *collegeService) CreateChecklist(ctx context.Context, userID string, createDTO *dto.StandaloneCreateChecklistDTO) (*models.Checklist, error) {
	collegeServiceLogger.Info("Creating checklist", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"name":           createDTO.Name,
	})

	// Validate DTO
	if err := s.validator.Struct(createDTO); err != nil {
		collegeServiceLogger.Error("Checklist validation failed", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify the application exists and user owns it
	_, err := s.GetByID(ctx, userID, createDTO.ApplicationID)
	if err != nil {
		return nil, err
	}

	// Create checklist
	checklist := &models.Checklist{
		ID:          primitive.NewObjectID(),
		Name:        createDTO.Name,
		Category:    createDTO.Category,
		Description: createDTO.Description,
		Tasks:       []models.Task{},
	}

	// Add checklist to application
	_, err = s.repo.PushToArray(ctx, createDTO.ApplicationID, "checklists", checklist)
	if err != nil {
		collegeServiceLogger.Error("Failed to create checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"error":          err.Error(),
		})
		return nil, err
	}

	collegeServiceLogger.Info("Checklist created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"checklist_id":   checklist.ID.Hex(),
	})

	return checklist, nil
}

func (s *collegeService) GetChecklist(ctx context.Context, userID string, applicationID string, checklistID string) (*models.Checklist, error) {
	// Verify the application exists and user owns it
	application, err := s.GetByID(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}

	// Find the checklist
	checklistObjectID, err := primitive.ObjectIDFromHex(checklistID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	for _, checklist := range application.Checklists {
		if checklist.ID == checklistObjectID {
			return &checklist, nil
		}
	}

	return nil, ErrNotFound
}

func (s *collegeService) UpdateChecklist(ctx context.Context, userID string, applicationID string, checklistID string, updateDTO *dto.StandaloneUpdateChecklistDTO) (*models.Checklist, error) {
	collegeServiceLogger.Info("Updating checklist", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
	})

	// Validate DTO
	if err := s.validator.Struct(updateDTO); err != nil {
		collegeServiceLogger.Error("Checklist update validation failed", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify checklist exists
	checklist, err := s.GetChecklist(ctx, userID, applicationID, checklistID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := bson.M{}
	if updateDTO.Name != nil {
		updates["checklists.$.name"] = *updateDTO.Name
	}
	if updateDTO.Category != nil {
		updates["checklists.$.category"] = *updateDTO.Category
	}
	if updateDTO.Description != nil {
		updates["checklists.$.description"] = *updateDTO.Description
	}

	if len(updates) == 0 {
		return checklist, nil
	}

	checklistObjectID, _ := primitive.ObjectIDFromHex(checklistID)
	applicationObjectID, _ := primitive.ObjectIDFromHex(applicationID)
	filter := bson.M{
		"_id":            applicationObjectID,
		"checklists._id": checklistObjectID,
	}

	_, err = s.repo.UpdateWithFilter(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		collegeServiceLogger.Error("Failed to update checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return nil, err
	}

	// Return updated checklist
	return s.GetChecklist(ctx, userID, applicationID, checklistID)
}

func (s *collegeService) DeleteChecklist(ctx context.Context, userID string, applicationID string, checklistID string) error {
	collegeServiceLogger.Info("Deleting checklist", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
	})

	// Verify checklist exists
	_, err := s.GetChecklist(ctx, userID, applicationID, checklistID)
	if err != nil {
		return err
	}

	checklistObjectID, _ := primitive.ObjectIDFromHex(checklistID)
	applicationObjectID, _ := primitive.ObjectIDFromHex(applicationID)
	filter := bson.M{"_id": applicationObjectID}
	pullUpdate := bson.M{
		"$pull": bson.M{"checklists": bson.M{"_id": checklistObjectID}},
	}

	_, err = s.repo.PullFromArray(ctx, filter, pullUpdate)
	if err != nil {
		collegeServiceLogger.Error("Failed to delete checklist", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"error":          err.Error(),
		})
		return err
	}

	collegeServiceLogger.Info("Checklist deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
	})

	return nil
}

// Task operations
func (s *collegeService) CreateTask(ctx context.Context, userID string, createDTO *dto.StandaloneCreateTaskDTO) (*models.Task, error) {
	collegeServiceLogger.Info("Creating task", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"checklist_id":   createDTO.ChecklistID,
		"name":           createDTO.Name,
	})

	// Validate DTO
	if err := s.validator.Struct(createDTO); err != nil {
		collegeServiceLogger.Error("Task validation failed", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"checklist_id":   createDTO.ChecklistID,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify checklist exists
	_, err := s.GetChecklist(ctx, userID, createDTO.ApplicationID, createDTO.ChecklistID)
	if err != nil {
		return nil, err
	}

	// Create task
	task := &models.Task{
		ID:          primitive.NewObjectID(),
		Completed:   false,
		Name:        createDTO.Name,
		Deadline:    createDTO.Deadline,
		Priority:    createDTO.Priority,
		Type:        createDTO.Type,
		Component:   createDTO.Component,
		Status:      createDTO.Status,
		Responsible: createDTO.Responsible,
		Notes:       createDTO.Notes,
	}

	checklistObjectID, err := primitive.ObjectIDFromHex(createDTO.ChecklistID)
	if err != nil {
		collegeServiceLogger.Error("Invalid checklist ID format", map[string]interface{}{
			"user_id":      userID,
			"checklist_id": createDTO.ChecklistID,
			"error":        err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	applicationObjectID, err := primitive.ObjectIDFromHex(createDTO.ApplicationID)
	if err != nil {
		collegeServiceLogger.Error("Invalid application ID format", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	filter := bson.M{
		"_id":            applicationObjectID,
		"checklists._id": checklistObjectID,
	}

	updates := bson.M{
		"$push": bson.M{"checklists.$.tasks": task},
	}

	_, err = s.repo.UpdateWithFilter(ctx, filter, updates)
	if err != nil {
		collegeServiceLogger.Error("Failed to create task", map[string]interface{}{
			"user_id":        userID,
			"application_id": createDTO.ApplicationID,
			"checklist_id":   createDTO.ChecklistID,
			"error":          err.Error(),
		})
		return nil, err
	}

	collegeServiceLogger.Info("Task created successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": createDTO.ApplicationID,
		"checklist_id":   createDTO.ChecklistID,
		"task_id":        task.ID.Hex(),
	})

	return task, nil
}

func (s *collegeService) GetTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string) (*models.Task, error) {
	// Get checklist first
	checklist, err := s.GetChecklist(ctx, userID, applicationID, checklistID)
	if err != nil {
		return nil, err
	}

	// Find the task
	taskObjectID, err := primitive.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, errors.ErrInvalidInput
	}

	for _, task := range checklist.Tasks {
		if task.ID == taskObjectID {
			return &task, nil
		}
	}

	return nil, ErrNotFound
}

func (s *collegeService) UpdateTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string, updateDTO *dto.StandaloneUpdateTaskDTO) (*models.Task, error) {
	collegeServiceLogger.Info("Updating task", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
		"task_id":        taskID,
	})

	// Validate DTO
	if err := s.validator.Struct(updateDTO); err != nil {
		collegeServiceLogger.Error("Task update validation failed", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return nil, errors.ErrInvalidInput
	}

	// Verify task exists
	task, err := s.GetTask(ctx, userID, applicationID, checklistID, taskID)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := bson.M{}
	if updateDTO.Completed != nil {
		updates["checklists.$[checklist].tasks.$[task].completed"] = *updateDTO.Completed
	}
	if updateDTO.Name != nil {
		updates["checklists.$[checklist].tasks.$[task].name"] = *updateDTO.Name
	}
	if updateDTO.Deadline != nil {
		updates["checklists.$[checklist].tasks.$[task].deadline"] = *updateDTO.Deadline
	}
	if updateDTO.Priority != nil {
		updates["checklists.$[checklist].tasks.$[task].priority"] = *updateDTO.Priority
	}
	if updateDTO.Type != nil {
		updates["checklists.$[checklist].tasks.$[task].type"] = *updateDTO.Type
	}
	if updateDTO.Component != nil {
		updates["checklists.$[checklist].tasks.$[task].component"] = *updateDTO.Component
	}
	if updateDTO.Status != nil {
		updates["checklists.$[checklist].tasks.$[task].status"] = *updateDTO.Status
	}
	if updateDTO.Responsible != nil {
		updates["checklists.$[checklist].tasks.$[task].responsible"] = *updateDTO.Responsible
	}
	if updateDTO.Notes != nil {
		updates["checklists.$[checklist].tasks.$[task].notes"] = *updateDTO.Notes
	}

	if len(updates) == 0 {
		return task, nil
	}

	checklistObjectID, _ := primitive.ObjectIDFromHex(checklistID)
	taskObjectID, _ := primitive.ObjectIDFromHex(taskID)
	applicationObjectID, _ := primitive.ObjectIDFromHex(applicationID)

	filter := bson.M{"_id": applicationObjectID}
	arrayFilters := []interface{}{
		bson.M{"checklist._id": checklistObjectID},
		bson.M{"task._id": taskObjectID},
	}

	_, err = s.repo.UpdateWithArrayFilters(ctx, filter, bson.M{"$set": updates}, arrayFilters)
	if err != nil {
		collegeServiceLogger.Error("Failed to update task", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return nil, err
	}

	// Return updated task
	return s.GetTask(ctx, userID, applicationID, checklistID, taskID)
}

func (s *collegeService) DeleteTask(ctx context.Context, userID string, applicationID string, checklistID string, taskID string) error {
	collegeServiceLogger.Info("Deleting task", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
		"task_id":        taskID,
	})

	// Verify task exists
	_, err := s.GetTask(ctx, userID, applicationID, checklistID, taskID)
	if err != nil {
		return err
	}

	checklistObjectID, _ := primitive.ObjectIDFromHex(checklistID)
	taskObjectID, _ := primitive.ObjectIDFromHex(taskID)
	applicationObjectID, _ := primitive.ObjectIDFromHex(applicationID)

	filter := bson.M{
		"_id":            applicationObjectID,
		"checklists._id": checklistObjectID,
	}

	pullUpdate := bson.M{
		"$pull": bson.M{"checklists.$.tasks": bson.M{"_id": taskObjectID}},
	}

	_, err = s.repo.PullFromArray(ctx, filter, pullUpdate)
	if err != nil {
		collegeServiceLogger.Error("Failed to delete task", map[string]interface{}{
			"user_id":        userID,
			"application_id": applicationID,
			"checklist_id":   checklistID,
			"task_id":        taskID,
			"error":          err.Error(),
		})
		return err
	}

	collegeServiceLogger.Info("Task deleted successfully", map[string]interface{}{
		"user_id":        userID,
		"application_id": applicationID,
		"checklist_id":   checklistID,
		"task_id":        taskID,
	})

	return nil
}
