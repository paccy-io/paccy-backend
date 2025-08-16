package dto

import (
	"time"

	"github.com/paccy-io/paccy-backend/internal/college/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateCollegeApplicationDTO struct {
	Name            string                 `json:"name" validate:"required,min=1,max=255"`
	Location        string                 `json:"location" validate:"omitempty,min=1,max=255"`
	Deadline        time.Time              `json:"deadline" validate:"omitempty"`     // optional
	Websites        []string               `json:"websites" validate:"dive,url"`      // optional
	Tags            []string               `json:"tags" validate:"dive,min=1,max=50"` // optional
	Notes           string                 `json:"notes" validate:"max=1000"`         //optional
	Status          models.Status          `json:"status" validate:"required,oneof=not_started in_progress submitted under_review accepted rejected deferred waitlisted withdrawn"`
	ApplicationType models.ApplicationType `json:"application_type" validate:"required,oneof=early_decision early_action regular_decision rolling_admission"`
	Priority        models.Priority        `json:"priority" validate:"required,oneof=reach target safety"`
	Major           string                 `json:"major" validate:"required,min=1,max=255"`
	ApplicationFee  int                    `json:"application_fee" validate:"min=0,max=10000"`
	ContactEmail    string                 `json:"contact_email" validate:"omitempty,email"`
	Checklists      []CreateChecklistDTO   `json:"checklists" validate:"dive"`
}

type UpdateCollegeApplicationDTO struct {
	Name            *string                 `json:"name" validate:"omitempty,min=1,max=255"`
	Location        *string                 `json:"location" validate:"omitempty,min=1,max=255"`
	Deadline        *time.Time              `json:"deadline"`
	Websites        []string                `json:"websites" validate:"dive,url"`
	Tags            []string                `json:"tags" validate:"dive,min=1,max=50"`
	Notes           *string                 `json:"notes" validate:"omitempty,max=1000"`
	Status          *models.Status          `json:"status" validate:"omitempty,oneof=not_started in_progress submitted under_review accepted rejected deferred waitlisted withdrawn"`
	ApplicationType *models.ApplicationType `json:"application_type" validate:"omitempty,oneof=early_decision early_action regular_decision rolling_admission"`
	Priority        *models.Priority        `json:"priority" validate:"omitempty,oneof=reach target safety"`
	Major           *string                 `json:"major" validate:"omitempty,min=1,max=255"`
	ApplicationFee  *int                    `json:"application_fee" validate:"omitempty,min=0,max=10000"`
	ContactEmail    *string                 `json:"contact_email" validate:"omitempty,email"`
	Checklists      []UpdateChecklistDTO    `json:"checklists" validate:"dive"`
}

type CreateChecklistDTO struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Category    models.Category `json:"category" validate:"required,oneof=application financial_aid scholarship housing academic personal other"`
	Description string          `json:"description" validate:"max=500"`
	Tasks       []CreateTaskDTO `json:"tasks" validate:"dive"`
}

type UpdateChecklistDTO struct {
	ID          *primitive.ObjectID `json:"id"`
	Name        *string             `json:"name" validate:"omitempty,min=1,max=255"`
	Category    *models.Category    `json:"category" validate:"omitempty,oneof=application financial_aid scholarship housing academic personal other"`
	Description *string             `json:"description" validate:"omitempty,max=500"`
	Tasks       []UpdateTaskDTO     `json:"tasks" validate:"dive"`
}

type CreateTaskDTO struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Deadline    time.Time              `json:"deadline" validate:"required"`
	Priority    models.TaskPriority    `json:"priority" validate:"required,oneof=high medium low"`
	Type        models.TaskType        `json:"type" validate:"required,oneof=document_preparation submission review follow_up research writing scheduling other"`
	Component   models.TaskComponent   `json:"component" validate:"required,oneof=common_app supplemental_essay academic_records standardized_tests supporting_documents financial_documents other"`
	Status      models.TaskStatus      `json:"status" validate:"required,oneof=not_started in_progress under_review completed waiting"`
	Responsible models.TaskResponsible `json:"responsible" validate:"required,oneof=student parent counselor teacher school_office other"`
	Notes       string                 `json:"notes" validate:"max=500"`
}

type UpdateTaskDTO struct {
	ID          *primitive.ObjectID     `json:"id"`
	Completed   *bool                   `json:"completed"`
	Name        *string                 `json:"name" validate:"omitempty,min=1,max=255"`
	Deadline    *time.Time              `json:"deadline"`
	Priority    *models.TaskPriority    `json:"priority" validate:"omitempty,oneof=high medium low"`
	Type        *models.TaskType        `json:"type" validate:"omitempty,oneof=document_preparation submission review follow_up research writing scheduling other"`
	Component   *models.TaskComponent   `json:"component" validate:"omitempty,oneof=common_app supplemental_essay academic_records standardized_tests supporting_documents financial_documents other"`
	Status      *models.TaskStatus      `json:"status" validate:"omitempty,oneof=not_started in_progress under_review completed waiting"`
	Responsible *models.TaskResponsible `json:"responsible" validate:"omitempty,oneof=student parent counselor teacher school_office other"`
	Notes       *string                 `json:"notes" validate:"omitempty,max=500"`
}

// Conversion methods to model types
func (dto *CreateCollegeApplicationDTO) ToModel(userID primitive.ObjectID) *models.CollegeApplication {
	application := &models.CollegeApplication{
		UserID:          userID,
		Name:            dto.Name,
		Location:        dto.Location,
		Deadline:        dto.Deadline,
		Websites:        dto.Websites,
		Tags:            dto.Tags,
		Notes:           dto.Notes,
		Status:          dto.Status,
		ApplicationType: dto.ApplicationType,
		Priority:        dto.Priority,
		Major:           dto.Major,
		ApplicationFee:  dto.ApplicationFee,
		ContactEmail:    dto.ContactEmail,
		Checklists:      make([]models.Checklist, len(dto.Checklists)),
	}

	for i, checklistDTO := range dto.Checklists {
		application.Checklists[i] = models.Checklist{
			ID:          primitive.NewObjectID(),
			Name:        checklistDTO.Name,
			Category:    checklistDTO.Category,
			Description: checklistDTO.Description,
			Tasks:       make([]models.Task, len(checklistDTO.Tasks)),
		}

		for j, taskDTO := range checklistDTO.Tasks {
			application.Checklists[i].Tasks[j] = models.Task{
				ID:          primitive.NewObjectID(),
				Completed:   false,
				Name:        taskDTO.Name,
				Deadline:    taskDTO.Deadline,
				Priority:    taskDTO.Priority,
				Type:        taskDTO.Type,
				Component:   taskDTO.Component,
				Status:      taskDTO.Status,
				Responsible: taskDTO.Responsible,
				Notes:       taskDTO.Notes,
			}
		}
	}

	return application
}

// Standalone DTOs for checklist operations
type StandaloneCreateChecklistDTO struct {
	ApplicationID string          `json:"application_id" validate:"required"`
	Name          string          `json:"name" validate:"required,min=1,max=255"`
	Category      models.Category `json:"category" validate:"required,oneof=application financial_aid scholarship housing academic personal other"`
	Description   string          `json:"description" validate:"max=500"`
}

type StandaloneUpdateChecklistDTO struct {
	Name        *string          `json:"name" validate:"omitempty,min=1,max=255"`
	Category    *models.Category `json:"category" validate:"omitempty,oneof=application financial_aid scholarship housing academic personal other"`
	Description *string          `json:"description" validate:"omitempty,max=500"`
}

// Standalone DTOs for task operations
type StandaloneCreateTaskDTO struct {
	ApplicationID string                 `json:"application_id" validate:"required"`
	ChecklistID   string                 `json:"checklist_id" validate:"required"`
	Name          string                 `json:"name" validate:"required,min=1,max=255"`
	Deadline      time.Time              `json:"deadline" validate:"required"`
	Priority      models.TaskPriority    `json:"priority" validate:"required,oneof=high medium low"`
	Type          models.TaskType        `json:"type" validate:"required,oneof=document_preparation submission review follow_up research writing scheduling other"`
	Component     models.TaskComponent   `json:"component" validate:"required,oneof=common_app supplemental_essay academic_records standardized_tests supporting_documents financial_documents other"`
	Status        models.TaskStatus      `json:"status" validate:"required,oneof=not_started in_progress under_review completed waiting"`
	Responsible   models.TaskResponsible `json:"responsible" validate:"required,oneof=student parent counselor teacher school_office other"`
	Notes         string                 `json:"notes" validate:"max=500"`
}

type StandaloneUpdateTaskDTO struct {
	Completed   *bool                   `json:"completed"`
	Name        *string                 `json:"name" validate:"omitempty,min=1,max=255"`
	Deadline    *time.Time              `json:"deadline"`
	Priority    *models.TaskPriority    `json:"priority" validate:"omitempty,oneof=high medium low"`
	Type        *models.TaskType        `json:"type" validate:"omitempty,oneof=document_preparation submission review follow_up research writing scheduling other"`
	Component   *models.TaskComponent   `json:"component" validate:"omitempty,oneof=common_app supplemental_essay academic_records standardized_tests supporting_documents financial_documents other"`
	Status      *models.TaskStatus      `json:"status" validate:"omitempty,oneof=not_started in_progress under_review completed waiting"`
	Responsible *models.TaskResponsible `json:"responsible" validate:"omitempty,oneof=student parent counselor teacher school_office other"`
	Notes       *string                 `json:"notes" validate:"omitempty,max=500"`
}
