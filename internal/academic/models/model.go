package models

import (
	. "github.com/paccy-io/paccy-backend/pkg/models"
)

type Exam struct {
	BaseModel `bson:",inline"`

	UserID   string   `bson:"user_id" json:"user_id"`
	Type     ExamType `bson:"type" json:"type"` // SAT, ACT, SAT Subject Test, AP, IB, TOEFL, IELTS, GRE, GMAT, LSAT, MCAT, other
	Name     string   `bson:"name" json:"name"`
	Score    string   `bson:"score" json:"score"`
	MaxScore string   `bson:"max_score" json:"max_score"`
	ExamDate string   `bson:"exam_date" json:"exam_date"` // YYYY-MM format
	Notes    string   `bson:"notes,omitempty" json:"notes,omitempty"`
}

type HighSchoolGPA struct {
	BaseModel `bson:",inline"`

	UserID string   `bson:"user_id" json:"user_id"`
	Scale  GPAScale `bson:"scale" json:"scale"`                     // Standardized scale types: 4.0, 5.0, 100
	GPA1   string   `bson:"gpa_1,omitempty" json:"gpa_1,omitempty"` // e.g., "3.8", "95" GRADE 9 - freshman (optional)
	GPA2   string   `bson:"gpa_2,omitempty" json:"gpa_2,omitempty"` // e.g., "3.9", "97" GRADE 10 - sophomore (optional)
	GPA3   string   `bson:"gpa_3,omitempty" json:"gpa_3,omitempty"` // e.g., "4.0", "100" GRADE 11 - junior (optional)
	GPA4   string   `bson:"gpa_4,omitempty" json:"gpa_4,omitempty"` // e.g., "3.7", "93" GRADE 12 - senior (optional)
	//OverallGPA string `bson:"overall_gpa" json:"overall_gpa"` // e.g., "3.85", "96.25" no need to store, can be calculated on the frontend
}

type ExamType string
type GPAScale string

const (
	ExamTypeSAT            ExamType = "SAT"
	ExamTypeACT            ExamType = "ACT"
	ExamTypeSATSubjectTest ExamType = "SAT Subject Test"
	ExamTypeAP             ExamType = "AP"
	ExamTypeIB             ExamType = "IB"
	ExamTypeTOEFL          ExamType = "TOEFL"
	ExamTypeIELTS          ExamType = "IELTS"
	ExamTypeGRE            ExamType = "GRE"
	ExamTypeGMAT           ExamType = "GMAT"
	ExamTypeLSAT           ExamType = "LSAT"
	ExamTypeMCAT           ExamType = "MCAT"
	ExamTypeOther          ExamType = "other"
)

const (
	GPAScale4Point0  GPAScale = "4.0" // Standard 4.0 scale (0.0-4.0)
	GPAScale5Point0  GPAScale = "5.0" // Weighted 5.0 scale (0.0-5.0)
	GPAScale100Point GPAScale = "100" // Percentage scale (0-100)
)

// Collection names
func GetExamCollectionName() string {
	return "exams"
}

func GetHighSchoolGPACollectionName() string {
	return "high_school_gpa"
}

// Lifecycle methods for Exam
func (e *Exam) BeforeCreate() {
	e.BaseModel.BeforeCreate()
}

func (e *Exam) BeforeUpdate() {
	e.BaseModel.BeforeUpdate()
}

// Lifecycle methods for HighSchoolGPA
func (h *HighSchoolGPA) BeforeCreate() {
	h.BaseModel.BeforeCreate()
}

func (h *HighSchoolGPA) BeforeUpdate() {
	h.BaseModel.BeforeUpdate()
}

// Validation function for ExamType
func IsValidExamType(examType string) ExamType {
	switch ExamType(examType) {
	case ExamTypeSAT, ExamTypeACT, ExamTypeSATSubjectTest, ExamTypeAP, ExamTypeIB,
		ExamTypeTOEFL, ExamTypeIELTS, ExamTypeGRE, ExamTypeGMAT, ExamTypeLSAT,
		ExamTypeMCAT, ExamTypeOther:
		return ExamType(examType)
	default:
		return ""
	}
}

// Validation function for GPAScale
func IsValidGPAScale(scale string) GPAScale {
	switch GPAScale(scale) {
	case GPAScale4Point0, GPAScale5Point0, GPAScale100Point:
		return GPAScale(scale)
	default:
		return ""
	}
}
