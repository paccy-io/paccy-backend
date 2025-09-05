package extracurriculars

import (
	"context"

	"github.com/paccy-io/paccy-backend/internal/extracurriculars/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var extracurricularLogger = logger.New("extracurricular")

type ExtracurricularRepository interface {
	Create(ctx context.Context, extracurricular *models.Extracurricular) (*models.Extracurricular, error)
	GetByID(ctx context.Context, id string) (*models.Extracurricular, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Extracurricular, error)
	GetByCategory(ctx context.Context, userID string, category string) ([]*models.Extracurricular, error)
	Update(ctx context.Context, id string, updates bson.M) (*models.Extracurricular, error)
	Delete(ctx context.Context, id string) error
}

type mongoExtracurricularRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
}

func NewMongoExtracurricularRepository(db *mongo.Database) ExtracurricularRepository {
	collection := db.Collection(models.GetCollectionName())
	return &mongoExtracurricularRepository{
		collection: collection,
		database:   db,
	}
}

func (r *mongoExtracurricularRepository) Create(ctx context.Context, extracurricular *models.Extracurricular) (*models.Extracurricular, error) {
	extracurricularLogger.Info("Creating new extracurricular", map[string]interface{}{
		"activity_name": extracurricular.ActivityName,
		"user_id":       extracurricular.UserID,
		"category":      string(extracurricular.Category),
	})

	extracurricular.BeforeCreate()

	result, err := r.collection.InsertOne(ctx, extracurricular)
	if err != nil {
		extracurricularLogger.Error("Failed to insert extracurricular", map[string]interface{}{
			"error":         err.Error(),
			"activity_name": extracurricular.ActivityName,
			"user_id":       extracurricular.UserID,
		})
		return nil, errors.ErrCreateFailed
	}

	extracurricular.ID = result.InsertedID.(primitive.ObjectID)
	extracurricularLogger.Info("Extracurricular created successfully", map[string]interface{}{
		"extracurricular_id": extracurricular.ID.Hex(),
		"activity_name":      extracurricular.ActivityName,
		"user_id":            extracurricular.UserID,
	})

	return extracurricular, nil
}

func (r *mongoExtracurricularRepository) GetByID(ctx context.Context, id string) (*models.Extracurricular, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		extracurricularLogger.Error("Invalid ObjectID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	var extracurricular models.Extracurricular
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&extracurricular)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			extracurricularLogger.Info("Extracurricular not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		extracurricularLogger.Error("Failed to get extracurricular", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return &extracurricular, nil
}

func (r *mongoExtracurricularRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Extracurricular, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": userID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		extracurricularLogger.Error("Failed to find extracurriculars for user", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var extracurriculars []*models.Extracurricular
	for cursor.Next(ctx) {
		var extracurricular models.Extracurricular
		if err := cursor.Decode(&extracurricular); err != nil {
			extracurricularLogger.Error("Failed to decode extracurricular", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		extracurriculars = append(extracurriculars, &extracurricular)
	}

	if err := cursor.Err(); err != nil {
		extracurricularLogger.Error("Cursor error while fetching extracurriculars", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	extracurricularLogger.Info("Retrieved extracurriculars for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(extracurriculars),
	})

	return extracurriculars, nil
}

func (r *mongoExtracurricularRepository) GetByCategory(ctx context.Context, userID string, category string) ([]*models.Extracurricular, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":  userID,
		"category": category,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		extracurricularLogger.Error("Failed to find extracurriculars by category", map[string]interface{}{
			"user_id":  userID,
			"category": category,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var extracurriculars []*models.Extracurricular
	for cursor.Next(ctx) {
		var extracurricular models.Extracurricular
		if err := cursor.Decode(&extracurricular); err != nil {
			extracurricularLogger.Error("Failed to decode extracurricular", map[string]interface{}{
				"user_id":  userID,
				"category": category,
				"error":    err.Error(),
			})
			continue
		}
		extracurriculars = append(extracurriculars, &extracurricular)
	}

	if err := cursor.Err(); err != nil {
		extracurricularLogger.Error("Cursor error while fetching extracurriculars by category", map[string]interface{}{
			"user_id":  userID,
			"category": category,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return extracurriculars, nil
}

func (r *mongoExtracurricularRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Extracurricular, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		extracurricularLogger.Error("Invalid ObjectID format for update", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var extracurricular models.Extracurricular
	if err := result.Decode(&extracurricular); err != nil {
		if err == mongo.ErrNoDocuments {
			extracurricularLogger.Warn("Extracurricular not found for update", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		extracurricularLogger.Error("Failed to update extracurricular", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	extracurricularLogger.Info("Extracurricular updated successfully", map[string]interface{}{
		"extracurricular_id": id,
		"activity_name":      extracurricular.ActivityName,
	})

	return &extracurricular, nil
}

func (r *mongoExtracurricularRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		extracurricularLogger.Error("Invalid ObjectID format for delete", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		extracurricularLogger.Error("Failed to delete extracurricular", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		extracurricularLogger.Warn("Extracurricular not found for deletion", map[string]interface{}{
			"id": id,
		})
		return errors.ErrGetFailed
	}

	extracurricularLogger.Info("Extracurricular deleted successfully", map[string]interface{}{
		"extracurricular_id": id,
	})

	return nil
}
