package models

import (
	"time"

	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollegeApplication struct {
	models.BaseModel `bson:",inline"` // make them in root level not in BaseModel, by using an inline struct

	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	Name            string             `json:"name" bson:"name"`
	Location        string             `json:"location" bson:"location"`
	Deadline        time.Time          `json:"deadline" bson:"deadline"`
	Websites        []string           `json:"websites" bson:"websites"`
	Tags            []string           `json:"tags" bson:"tags"`   // e.g. Ivy League,
	Notes           string             `json:"notes" bson:"notes"` // user's notes
	Status          Status             `json:"status" bson:"status"`
	ApplicationType ApplicationType    `json:"application_type" bson:"application_type"`
	Priority        Priority           `json:"priority" bson:"priority"` // e.g. reach, match (=target), safety
	Major           string             `json:"major" bson:"major"`
	ApplicationFee  int                `json:"application_fee" bson:"application_fee"`
	ContactEmail    string             `json:"contact_email" bson:"contact_email"`
	Checklists      []Checklist        `json:"checklists" bson:"checklists"`
}

type Checklist struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`    // unique identifier for the checklist
	Name        string             `json:"name" bson:"name"` // e.g. Application Material, Financial Aid, Supplemental Essays
	Category    Category           `json:"category" bson:"category"`
	Description string             `json:"description" bson:"description"`
	Tasks       []Task             `json:"tasks" bson:"tasks"` // e.g. list of tasks to complete
}

type Task struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"` // unique identifier for the task
	Completed   bool               `json:"completed" bson:"completed"`
	Name        string             `json:"name" bson:"name"` // e.g. Write personal statement
	Deadline    time.Time          `json:"deadline" bson:"deadline"`
	Priority    TaskPriority       `json:"priority" bson:"priority"`
	Type        TaskType           `json:"type" bson:"type"`
	Component   TaskComponent      `json:"component" bson:"component"`
	Status      TaskStatus         `json:"status" bson:"status"`
	Responsible TaskResponsible    `json:"responsible" bson:"responsible"`
	Notes       string             `json:"notes" bson:"notes"` // e.g. Additional notes or instructions
}

func (ca *CollegeApplication) BeforeCreate() error {
	ca.BaseModel.BeforeCreate()

	if ca.UserID.IsZero() {
		return ErrMissingUserID
	}

	if ca.Name == "" {
		return ErrMissingName
	}

	return nil
}

func (ca *CollegeApplication) BeforeUpdate() {
	ca.BaseModel.BeforeUpdate()
}

func (ca *CollegeApplication) BeforeDelete() {
	ca.BaseModel.BeforeDelete()
}
