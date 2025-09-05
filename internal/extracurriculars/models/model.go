package models

import (
	. "github.com/paccy-io/paccy-backend/pkg/models"
)

type Extracurricular struct {
	BaseModel `bson:",inline"`

	UserID       string   `bson:"user_id" json:"user_id"`
	ActivityName string   `bson:"activity_name" json:"activity_name"`
	Category     Category `bson:"category" json:"category"`
	Position     string   `bson:"position" json:"position"`
	HoursPerWeek int      `bson:"hours_per_week" json:"hours_per_week"`
	StartDate    string   `bson:"start_date" json:"start_date"`                         // YYYY-MM format
	EndDate      string   `bson:"end_date" json:"end_date"`                             // YYYY-MM format, or "present" if ongoing
	Description  string   `bson:"description,omitempty" json:"description,omitempty"`   // optional
	Location     string   `bson:"location,omitempty" json:"location,omitempty"`         // optional
	Website      string   `bson:"website,omitempty" json:"website,omitempty"`           // optional
	AIScore      AIScore  `bson:"ai_score,omitempty" json:"ai_score,omitempty"`         // high, medium, low, or empty string if not evaluated by AI
	Achievements []string `bson:"achievements,omitempty" json:"achievements,omitempty"` // optional, list of notable achievements
	Skills       []string `bson:"skills,omitempty" json:"skills,omitempty"`             // optional, list of skills gained

}

type Category string

// academic, sports, music, arts, community_service, leadership, technology, volunteering, clubs, work_experience, research, other
const (
	CategoryAcademic         Category = "academic"
	CategorySports           Category = "sports"
	CategoryMusic            Category = "music"
	CategoryArts             Category = "arts"
	CategoryCommunityService Category = "community_service"
	CategoryLeadership       Category = "leadership"
	CategoryTechnology       Category = "technology"
	CategoryVolunteering     Category = "volunteering"
	CategoryClubs            Category = "clubs"
	CategoryWorkExperience   Category = "work_experience"
	CategoryResearch         Category = "research"
	CategoryOther            Category = "other"
)

type AIScore string

const (
	AIScoreHigh   AIScore = "high"
	AIScoreMedium AIScore = "medium"
	AIScoreLow    AIScore = "low"
)

func GetCollectionName() string {
	return "extracurriculars"
}

func (e *Extracurricular) BeforeCreate() {
	e.BaseModel.BeforeCreate()
}

func (e *Extracurricular) BeforeUpdate() {
	e.BaseModel.BeforeUpdate()
}

func IsValidCategory(category string) Category {
	switch Category(category) {
	case CategoryAcademic, CategorySports, CategoryMusic, CategoryArts, CategoryCommunityService,
		CategoryLeadership, CategoryTechnology, CategoryVolunteering, CategoryClubs,
		CategoryWorkExperience, CategoryResearch, CategoryOther:
		return Category(category)
	default:
		return ""
	}
}
