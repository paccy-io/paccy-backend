package academic

import (
	"context"
	"time"

	"github.com/paccy-io/paccy-backend/internal/academic/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var repoLogger = logger.New("academic")

type AcademicRepository interface {
	// Exam methods
	CreateExam(ctx context.Context, exam *models.Exam) (*models.Exam, error)
	GetExamByID(ctx context.Context, id string) (*models.Exam, error)
	GetExamsByUserID(ctx context.Context, userID string) ([]*models.Exam, error)
	GetExamsByType(ctx context.Context, userID string, examType string) ([]*models.Exam, error)
	UpdateExam(ctx context.Context, id string, updates map[string]interface{}) (*models.Exam, error)
	DeleteExam(ctx context.Context, id string) error

	// HighSchoolGPA methods
	CreateHighSchoolGPA(ctx context.Context, gpa *models.HighSchoolGPA) (*models.HighSchoolGPA, error)
	GetHighSchoolGPAByUserID(ctx context.Context, userID string) (*models.HighSchoolGPA, error)
	UpdateHighSchoolGPA(ctx context.Context, id string, updates map[string]interface{}) (*models.HighSchoolGPA, error)
	DeleteHighSchoolGPA(ctx context.Context, id string) error
}

type academicRepository struct {
	db             *mongo.Database
	examCollection *mongo.Collection
	gpaCollection  *mongo.Collection
}

func NewAcademicRepository(db *mongo.Database) AcademicRepository {
	return &academicRepository{
		db:             db,
		examCollection: db.Collection(models.GetExamCollectionName()),
		gpaCollection:  db.Collection(models.GetHighSchoolGPACollectionName()),
	}
}

// Exam methods
func (r *academicRepository) CreateExam(ctx context.Context, exam *models.Exam) (*models.Exam, error) {
	repoLogger.Info("Creating new exam", map[string]interface{}{
		"exam_name": exam.Name,
		"exam_type": exam.Type,
		"user_id":   exam.UserID,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.examCollection.InsertOne(ctx, exam)
	if err != nil {
		repoLogger.Error("Failed to create exam", map[string]interface{}{
			"error":     err.Error(),
			"exam_name": exam.Name,
			"user_id":   exam.UserID,
		})
		return nil, errors.ErrCreateFailed
	}

	exam.ID = result.InsertedID.(primitive.ObjectID)

	repoLogger.Info("Exam created successfully", map[string]interface{}{
		"exam_id":   exam.ID.Hex(),
		"exam_name": exam.Name,
		"user_id":   exam.UserID,
	})

	return exam, nil
}

func (r *academicRepository) GetExamByID(ctx context.Context, id string) (*models.Exam, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repoLogger.Error("Invalid exam ID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var exam models.Exam
	err = r.examCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&exam)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			repoLogger.Info("Exam not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		repoLogger.Error("Failed to get exam by ID", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, errors.ErrGetFailed
	}

	return &exam, nil
}

func (r *academicRepository) GetExamsByUserID(ctx context.Context, userID string) ([]*models.Exam, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.examCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		repoLogger.Error("Failed to get exams by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var exams []*models.Exam
	if err = cursor.All(ctx, &exams); err != nil {
		repoLogger.Error("Failed to decode exams", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, errors.ErrGetFailed
	}

	repoLogger.Info("Retrieved exams for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(exams),
	})

	return exams, nil
}

func (r *academicRepository) GetExamsByType(ctx context.Context, userID string, examType string) ([]*models.Exam, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.examCollection.Find(ctx, bson.M{
		"user_id": userID,
		"type":    examType,
	})
	if err != nil {
		repoLogger.Error("Failed to get exams by type", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID,
			"exam_type": examType,
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var exams []*models.Exam
	if err = cursor.All(ctx, &exams); err != nil {
		repoLogger.Error("Failed to decode exams", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID,
			"exam_type": examType,
		})
		return nil, errors.ErrGetFailed
	}

	repoLogger.Info("Retrieved exams by type for user", map[string]interface{}{
		"user_id":   userID,
		"exam_type": examType,
		"count":     len(exams),
	})

	return exams, nil
}

func (r *academicRepository) UpdateExam(ctx context.Context, id string, updates map[string]interface{}) (*models.Exam, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repoLogger.Error("Invalid exam ID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now()

	_, err = r.examCollection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		repoLogger.Error("Failed to update exam", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, errors.ErrUpdateFailed
	}

	// Get updated exam
	updatedExam, err := r.GetExamByID(ctx, id)
	if err != nil {
		return nil, err
	}

	repoLogger.Info("Exam updated successfully", map[string]interface{}{
		"exam_id":   id,
		"exam_name": updatedExam.Name,
	})

	return updatedExam, nil
}

func (r *academicRepository) DeleteExam(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repoLogger.Error("Invalid exam ID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.examCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		repoLogger.Error("Failed to delete exam", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		repoLogger.Info("Exam not found for deletion", map[string]interface{}{
			"id": id,
		})
		return errors.ErrGetFailed
	}

	repoLogger.Info("Exam deleted successfully", map[string]interface{}{
		"exam_id": id,
	})

	return nil
}

// HighSchoolGPA methods
func (r *academicRepository) CreateHighSchoolGPA(ctx context.Context, gpa *models.HighSchoolGPA) (*models.HighSchoolGPA, error) {
	repoLogger.Info("Creating new HighSchoolGPA", map[string]interface{}{
		"user_id": gpa.UserID,
		"scale":   gpa.Scale,
	})

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.gpaCollection.InsertOne(ctx, gpa)
	if err != nil {
		repoLogger.Error("Failed to create HighSchoolGPA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": gpa.UserID,
		})
		return nil, errors.ErrCreateFailed
	}

	gpa.ID = result.InsertedID.(primitive.ObjectID)

	repoLogger.Info("HighSchoolGPA created successfully", map[string]interface{}{
		"gpa_id":  gpa.ID.Hex(),
		"user_id": gpa.UserID,
	})

	return gpa, nil
}

func (r *academicRepository) GetHighSchoolGPAByUserID(ctx context.Context, userID string) (*models.HighSchoolGPA, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var gpa models.HighSchoolGPA
	err := r.gpaCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&gpa)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			repoLogger.Info("HighSchoolGPA not found", map[string]interface{}{
				"user_id": userID,
			})
			return nil, errors.ErrGetFailed
		}
		repoLogger.Error("Failed to get HighSchoolGPA by user ID", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return nil, errors.ErrGetFailed
	}

	return &gpa, nil
}

func (r *academicRepository) UpdateHighSchoolGPA(ctx context.Context, id string, updates map[string]interface{}) (*models.HighSchoolGPA, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repoLogger.Error("Invalid HighSchoolGPA ID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now()

	_, err = r.gpaCollection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	if err != nil {
		repoLogger.Error("Failed to update HighSchoolGPA", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, errors.ErrUpdateFailed
	}

	// Get updated GPA
	var updatedGPA models.HighSchoolGPA
	err = r.gpaCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updatedGPA)
	if err != nil {
		repoLogger.Error("Failed to get updated HighSchoolGPA", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, errors.ErrGetFailed
	}

	repoLogger.Info("HighSchoolGPA updated successfully", map[string]interface{}{
		"gpa_id": id,
	})

	return &updatedGPA, nil
}

func (r *academicRepository) DeleteHighSchoolGPA(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repoLogger.Error("Invalid HighSchoolGPA ID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := r.gpaCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		repoLogger.Error("Failed to delete HighSchoolGPA", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		repoLogger.Info("HighSchoolGPA not found for deletion", map[string]interface{}{
			"id": id,
		})
		return errors.ErrGetFailed
	}

	repoLogger.Info("HighSchoolGPA deleted successfully", map[string]interface{}{
		"gpa_id": id,
	})

	return nil
}
