package essay

import (
	"context"
	"time"

	"github.com/paccy-io/paccy-backend/internal/essay/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var essayLogger = logger.New("essay")

type EssayRepository interface {
	Create(ctx context.Context, essay *models.Essay) (*models.Essay, error)
	GetByID(ctx context.Context, id string) (*models.Essay, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Essay, error)
	GetByCollegeApplicationID(ctx context.Context, collegeApplicationID string) ([]*models.Essay, error)
	Update(ctx context.Context, id string, updates bson.M) (*models.Essay, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, userID string) (int64, error)
	GetByStatus(ctx context.Context, userID string, status string) ([]*models.Essay, error)
	GetEssaysWithDeadlines(ctx context.Context, userID string, days int) ([]*models.Essay, error)
}

type mongoEssayRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
}

func NewMongoEssayRepository(db *mongo.Database) EssayRepository {
	collection := db.Collection(models.GetCollectionName())
	return &mongoEssayRepository{
		collection: collection,
		database:   db,
	}
}

func (r *mongoEssayRepository) Create(ctx context.Context, essay *models.Essay) (*models.Essay, error) {
	essayLogger.Info("Creating new essay", map[string]interface{}{
		"title":                  essay.Title,
		"college_application_id": essay.CollegeApplicationID.Hex(),
	})

	essay.BeforeCreate()

	result, err := r.collection.InsertOne(ctx, essay)
	if err != nil {
		essayLogger.Error("Failed to insert essay", map[string]interface{}{
			"error": err.Error(),
			"title": essay.Title,
		})
		return nil, errors.ErrCreateFailed
	}

	essay.ID = result.InsertedID.(primitive.ObjectID)
	essayLogger.Info("Essay created successfully", map[string]interface{}{
		"essay_id": essay.ID.Hex(),
		"title":    essay.Title,
	})

	return essay, nil
}

func (r *mongoEssayRepository) GetByID(ctx context.Context, id string) (*models.Essay, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		essayLogger.Error("Invalid ObjectID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	var essay models.Essay
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&essay)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			essayLogger.Info("Essay not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		essayLogger.Error("Failed to get essay", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return &essay, nil
}

func (r *mongoEssayRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Essay, error) {
	// We need to use aggregation pipeline to join with college_applications to get essays for a user
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		essayLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	pipeline := []bson.M{
		// Join with college_applications collection
		{
			"$lookup": bson.M{
				"from":         "college_applications",
				"localField":   "college_application_id",
				"foreignField": "_id",
				"as":           "college_application",
			},
		},
		// Match essays where college application belongs to the user
		{
			"$match": bson.M{
				"college_application.user_id": userObjectID,
			},
		},
		// Remove the joined college_application field to return clean essay documents
		{
			"$unset": "college_application",
		},
		// Sort by created_at descending
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		essayLogger.Error("Failed to find essays for user", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var essays []*models.Essay
	for cursor.Next(ctx) {
		var essay models.Essay
		if err := cursor.Decode(&essay); err != nil {
			essayLogger.Error("Failed to decode essay", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		essays = append(essays, &essay)
	}

	if err := cursor.Err(); err != nil {
		essayLogger.Error("Cursor error while fetching essays", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	essayLogger.Info("Retrieved essays for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(essays),
	})

	return essays, nil
}

func (r *mongoEssayRepository) GetByCollegeApplicationID(ctx context.Context, collegeApplicationID string) ([]*models.Essay, error) {
	objectID, err := primitive.ObjectIDFromHex(collegeApplicationID)
	if err != nil {
		essayLogger.Error("Invalid college application ObjectID format", map[string]interface{}{
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"college_application_id": objectID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		essayLogger.Error("Failed to find essays for college application", map[string]interface{}{
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var essays []*models.Essay
	for cursor.Next(ctx) {
		var essay models.Essay
		if err := cursor.Decode(&essay); err != nil {
			essayLogger.Error("Failed to decode essay", map[string]interface{}{
				"college_application_id": collegeApplicationID,
				"error":                  err.Error(),
			})
			continue
		}
		essays = append(essays, &essay)
	}

	if err := cursor.Err(); err != nil {
		essayLogger.Error("Cursor error while fetching essays", map[string]interface{}{
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return essays, nil
}

func (r *mongoEssayRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Essay, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		essayLogger.Error("Invalid ObjectID format for update", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	// Add updated_at timestamp
	updates["updated_at"] = time.Now()

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var essay models.Essay
	if err := result.Decode(&essay); err != nil {
		if err == mongo.ErrNoDocuments {
			essayLogger.Warn("Essay not found for update", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		essayLogger.Error("Failed to update essay", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	essayLogger.Info("Essay updated successfully", map[string]interface{}{
		"essay_id": id,
		"title":    essay.Title,
	})

	return &essay, nil
}

func (r *mongoEssayRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		essayLogger.Error("Invalid ObjectID format for delete", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	// Hard delete: permanently remove the document
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		essayLogger.Error("Failed to delete essay", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		essayLogger.Warn("Essay not found for deletion", map[string]interface{}{
			"id": id,
		})
		return errors.ErrGetFailed
	}

	essayLogger.Info("Essay deleted successfully", map[string]interface{}{
		"essay_id": id,
	})

	return nil
}

func (r *mongoEssayRepository) Count(ctx context.Context, userID string) (int64, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		essayLogger.Error("Invalid user ObjectID format for count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrInvalidObjectID
	}

	// Use aggregation to count essays belonging to user's college applications
	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "college_applications",
				"localField":   "college_application_id",
				"foreignField": "_id",
				"as":           "college_application",
			},
		},
		{
			"$match": bson.M{
				"college_application.user_id": userObjectID,
			},
		},
		{
			"$count": "total",
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		essayLogger.Error("Failed to count essays", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, errors.ErrGetFailed
	}

	if len(result) == 0 {
		return 0, nil
	}

	total, ok := result[0]["total"].(int32)
	if !ok {
		return 0, nil
	}

	return int64(total), nil
}

func (r *mongoEssayRepository) GetByStatus(ctx context.Context, userID string, status string) ([]*models.Essay, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		essayLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "college_applications",
				"localField":   "college_application_id",
				"foreignField": "_id",
				"as":           "college_application",
			},
		},
		{
			"$match": bson.M{
				"college_application.user_id": userObjectID,
				"status":                      status,
			},
		},
		{
			"$unset": "college_application",
		},
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		essayLogger.Error("Failed to find essays by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var essays []*models.Essay
	for cursor.Next(ctx) {
		var essay models.Essay
		if err := cursor.Decode(&essay); err != nil {
			essayLogger.Error("Failed to decode essay", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		essays = append(essays, &essay)
	}

	if err := cursor.Err(); err != nil {
		essayLogger.Error("Cursor error while fetching essays by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return essays, nil
}

func (r *mongoEssayRepository) GetEssaysWithDeadlines(ctx context.Context, userID string, days int) ([]*models.Essay, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		essayLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	pipeline := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "college_applications",
				"localField":   "college_application_id",
				"foreignField": "_id",
				"as":           "college_application",
			},
		},
		{
			"$match": bson.M{
				"college_application.user_id": userObjectID,
				"college_application.deadline": bson.M{
					"$gte": now,
					"$lte": endDate,
				},
			},
		},
		{
			"$unset": "college_application",
		},
		{
			"$sort": bson.M{"college_application.deadline": 1}, // Sort by deadline ascending
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		essayLogger.Error("Failed to find essays with upcoming deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var essays []*models.Essay
	for cursor.Next(ctx) {
		var essay models.Essay
		if err := cursor.Decode(&essay); err != nil {
			essayLogger.Error("Failed to decode essay", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		essays = append(essays, &essay)
	}

	if err := cursor.Err(); err != nil {
		essayLogger.Error("Cursor error while fetching essays with deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	essayLogger.Info("Retrieved essays with upcoming deadlines", map[string]interface{}{
		"user_id": userID,
		"days":    days,
		"count":   len(essays),
	})

	return essays, nil
}
