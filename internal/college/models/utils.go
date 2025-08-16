package models

type Status string

const (
	StatusNotStarted     Status = "not_started"
	StatusInProgress     Status = "in_progress"
	StatusSubmitted      Status = "submitted"
	StatusUnderReview    Status = "under_review"
	StatusAccepted       Status = "accepted"
	StatusRejected       Status = "rejected"
	StatusDeferred       Status = "deferred"
	StatusWaitlisted     Status = "waitlisted"
	StatusWithdrawn      Status = "withdrawn"
)

type Priority string

const (
	PriorityReach  Priority = "reach"
	PriorityTarget Priority = "target"
	PrioritySafety  Priority = "safety"
)

type ApplicationType string

const (
	ApplicationTypeEarlyDecision      ApplicationType = "early_decision"
	ApplicationTypeEarlyAction        ApplicationType = "early_action"
	ApplicationTypeRegularDecision    ApplicationType = "regular_decision"
	ApplicationTypeRollingAdmission   ApplicationType = "rolling_admission"
)

type Category string

const (
	CategoryApplication      Category = "application"
	CategoryFinancialAid     Category = "financial_aid"
	CategoryScholarship      Category = "scholarship"
	CategoryHousing          Category = "housing"
	CategoryAcademic         Category = "academic"
	CategoryPersonal        Category = "personal"
	CategoryOther           Category = "other"
)

type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)


type TaskType string

const (
	TaskTypeDocumentPreparation TaskType = "document_preparation"
	TaskTypeSubmission           TaskType = "submission"
	TaskTypeReview               TaskType = "review"
	TaskTypeFollowUp             TaskType = "follow_up"
	TaskTypeResearch             TaskType = "research"
	TaskTypeWriting              TaskType = "writing"
	TaskTypeScheduling           TaskType = "scheduling"
	TaskTypeOther                TaskType = "other"
)

type TaskComponent string

const (
	TaskComponentCommonApp          TaskComponent = "common_app"
	TaskComponentSupplementalEssay  TaskComponent = "supplemental_essay"
	TaskComponentAcademicRecords    TaskComponent = "academic_records"
	TaskComponentStandardizedTests   TaskComponent = "standardized_tests"
	TaskComponentSupportingDocuments TaskComponent = "supporting_documents"
	TaskComponentFinancialDocuments  TaskComponent = "financial_documents"
	TaskComponentOther               TaskComponent = "other"
)

type TaskStatus string

const (
	TaskStatusNotStarted     TaskStatus = "not_started"
	TaskStatusInProgress     TaskStatus = "in_progress"
	TaskStatusUnderReview    TaskStatus = "under_review"
	TaskStatusCompleted      TaskStatus = "completed"
	TaskStatusWaiting        TaskStatus = "waiting"
)

type TaskResponsible string

const (
	TaskResponsibleStudent   TaskResponsible = "student"
	TaskResponsibleParent    TaskResponsible = "parent"
	TaskResponsibleCounselor TaskResponsible = "counselor"
	TaskResponsibleTeacher   TaskResponsible = "teacher"
	TaskResponsibleSchoolOffice TaskResponsible = "school_office"
	TaskResponsibleOther     TaskResponsible = "other"
)
