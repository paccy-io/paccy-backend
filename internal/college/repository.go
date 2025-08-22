package college

import (
	"context"
	"time"

	"github.com/paccy-io/paccy-backend/internal/college/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collegeLogger = logger.New("college")

type CollegeRepository interface {
	Create(ctx context.Context, application *models.CollegeApplication) (*models.CollegeApplication, error)
	GetByID(ctx context.Context, id string) (*models.CollegeApplication, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.CollegeApplication, error)
	Update(ctx context.Context, id string, updates bson.M) (*models.CollegeApplication, error)
	UpdateWithFilter(ctx context.Context, filter bson.M, updates bson.M) (*models.CollegeApplication, error)
	UpdateWithArrayFilters(ctx context.Context, filter bson.M, updates bson.M, arrayFilters []interface{}) (*models.CollegeApplication, error)
	PushToArray(ctx context.Context, id string, arrayField string, item interface{}) (*models.CollegeApplication, error)
	PullFromArray(ctx context.Context, filter bson.M, pullUpdate bson.M) (*models.CollegeApplication, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, userID string) (int64, error)
	GetByStatus(ctx context.Context, userID string, status string) ([]*models.CollegeApplication, error)
	GetByPriority(ctx context.Context, userID string, priority string) ([]*models.CollegeApplication, error)
	GetUpcomingDeadlines(ctx context.Context, userID string, days int) ([]*models.CollegeApplication, error)
}

type mongoCollegeRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
}

func NewMongoCollegeRepository(db *mongo.Database) CollegeRepository {
	collection := db.Collection("college_applications")
	return &mongoCollegeRepository{
		collection: collection,
		database:   db,
	}
}

func (r *mongoCollegeRepository) Create(ctx context.Context, application *models.CollegeApplication) (*models.CollegeApplication, error) {
	collegeLogger.Info("Creating new college application", map[string]interface{}{
		"user_id": application.UserID.Hex(),
		"name":    application.Name,
	})

	if err := application.BeforeCreate(); err != nil {
		collegeLogger.Error("Validation failed during college application creation", map[string]interface{}{
			"error":   err.Error(),
			"user_id": application.UserID.Hex(),
		})
		return nil, err
	}

	result, err := r.collection.InsertOne(ctx, application)
	if err != nil {
		collegeLogger.Error("Failed to insert college application", map[string]interface{}{
			"error":   err.Error(),
			"user_id": application.UserID.Hex(),
		})
		return nil, errors.ErrCreateFailed
	}

	application.ID = result.InsertedID.(primitive.ObjectID)
	collegeLogger.Info("College application created successfully", map[string]interface{}{
		"application_id": application.ID.Hex(),
		"user_id":        application.UserID.Hex(),
	})

	return application, nil
}

func (r *mongoCollegeRepository) GetByID(ctx context.Context, id string) (*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		collegeLogger.Error("Invalid ObjectID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	var application models.CollegeApplication
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&application)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Info("College application not found for deletion", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to get college application", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return &application, nil
}

func (r *mongoCollegeRepository) GetByUserID(ctx context.Context, userID string) ([]*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objectID})
	if err != nil {
		collegeLogger.Error("Failed to find college applications for user", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var applications []*models.CollegeApplication
	for cursor.Next(ctx) {
		var application models.CollegeApplication
		if err := cursor.Decode(&application); err != nil {
			collegeLogger.Error("Failed to decode college application", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		applications = append(applications, &application)
	}

	if err := cursor.Err(); err != nil {
		collegeLogger.Error("Cursor error while fetching college applications", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	collegeLogger.Info("Retrieved college applications for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(applications),
	})

	return applications, nil
}

func (r *mongoCollegeRepository) Update(ctx context.Context, id string, updates bson.M) (*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		collegeLogger.Error("Invalid ObjectID format for update", map[string]interface{}{
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

	var application models.CollegeApplication
	if err := result.Decode(&application); err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Warn("College application not found for update", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to update college application", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	collegeLogger.Info("College application updated successfully", map[string]interface{}{
		"application_id": id,
		"user_id":        application.UserID.Hex(),
	})

	return &application, nil
}

func (r *mongoCollegeRepository) UpdateWithFilter(ctx context.Context, filter bson.M, updates bson.M) (*models.CollegeApplication, error) {
	// Don't modify the updates here since it might contain MongoDB operators
	// Just add the updated_at field if we have a $set operator
	if _, hasSet := updates["$set"]; hasSet {
		updates["$set"].(bson.M)["updated_at"] = time.Now()
	} else if len(updates) > 0 && !containsMongoOperators(updates) {
		// Only add updated_at if this is a simple field update (no operators)
		updates["updated_at"] = time.Now()
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		updates,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var application models.CollegeApplication
	if err := result.Decode(&application); err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Warn("College application not found for update with filter", map[string]interface{}{
				"filter": filter,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to update college application with filter", map[string]interface{}{
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	collegeLogger.Info("College application updated successfully with filter", map[string]interface{}{
		"application_id": application.ID.Hex(),
		"user_id":        application.UserID.Hex(),
	})

	return &application, nil
}

// Helper function to check if updates contain MongoDB operators
func containsMongoOperators(updates bson.M) bool {
	for key := range updates {
		if len(key) > 0 && key[0] == '$' {
			return true
		}
	}
	return false
}

func (r *mongoCollegeRepository) UpdateWithArrayFilters(ctx context.Context, filter bson.M, updates bson.M, arrayFilters []interface{}) (*models.CollegeApplication, error) {
	// Add updated_at timestamp
	if setUpdates, ok := updates["$set"].(bson.M); ok {
		setUpdates["updated_at"] = time.Now()
	} else {
		if updates["$set"] == nil {
			updates["$set"] = bson.M{}
		}
		updates["$set"].(bson.M)["updated_at"] = time.Now()
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetArrayFilters(options.ArrayFilters{Filters: arrayFilters})

	result := r.collection.FindOneAndUpdate(ctx, filter, updates, opts)

	var application models.CollegeApplication
	if err := result.Decode(&application); err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Warn("College application not found for update with array filters", map[string]interface{}{
				"filter": filter,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to update college application with array filters", map[string]interface{}{
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	collegeLogger.Info("College application updated successfully with array filters", map[string]interface{}{
		"application_id": application.ID.Hex(),
		"user_id":        application.UserID.Hex(),
	})

	return &application, nil
}

func (r *mongoCollegeRepository) PushToArray(ctx context.Context, id string, arrayField string, item interface{}) (*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		collegeLogger.Error("Invalid ObjectID format for push operation", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	updates := bson.M{
		"$push": bson.M{arrayField: item},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		updates,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var application models.CollegeApplication
	if err := result.Decode(&application); err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Warn("College application not found for push operation", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to push to array in college application", map[string]interface{}{
			"id":    id,
			"field": arrayField,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	collegeLogger.Info("Successfully pushed to array in college application", map[string]interface{}{
		"application_id": application.ID.Hex(),
		"user_id":        application.UserID.Hex(),
		"field":          arrayField,
	})

	return &application, nil
}

func (r *mongoCollegeRepository) PullFromArray(ctx context.Context, filter bson.M, pullUpdate bson.M) (*models.CollegeApplication, error) {
	// Add updated_at timestamp
	pullUpdate["$set"] = bson.M{"updated_at": time.Now()}

	result := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		pullUpdate,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var application models.CollegeApplication
	if err := result.Decode(&application); err != nil {
		if err == mongo.ErrNoDocuments {
			collegeLogger.Warn("College application not found for pull operation", map[string]interface{}{
				"filter": filter,
			})
			return nil, errors.ErrCollegeNotFound
		}
		collegeLogger.Error("Failed to pull from array in college application", map[string]interface{}{
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	collegeLogger.Info("Successfully pulled from array in college application", map[string]interface{}{
		"application_id": application.ID.Hex(),
		"user_id":        application.UserID.Hex(),
	})

	return &application, nil
}

func (r *mongoCollegeRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		collegeLogger.Error("Invalid ObjectID format for delete", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		collegeLogger.Error("Failed to delete college application", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		collegeLogger.Warn("College application not found for deletion", map[string]interface{}{
			"id": id,
		})
		return ErrNotFound
	}

	collegeLogger.Info("College application deleted successfully", map[string]interface{}{
		"application_id": id,
	})

	return nil
}

func (r *mongoCollegeRepository) Count(ctx context.Context, userID string) (int64, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeLogger.Error("Invalid user ObjectID format for count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrInvalidObjectID
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": objectID})
	if err != nil {
		collegeLogger.Error("Failed to count college applications", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrGetFailed
	}

	return count, nil
}

func (r *mongoCollegeRepository) GetByStatus(ctx context.Context, userID string, status string) ([]*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	filter := bson.M{
		"user_id": objectID,
		"status":  status,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		collegeLogger.Error("Failed to find college applications by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var applications []*models.CollegeApplication
	for cursor.Next(ctx) {
		var application models.CollegeApplication
		if err := cursor.Decode(&application); err != nil {
			collegeLogger.Error("Failed to decode college application", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		applications = append(applications, &application)
	}

	if err := cursor.Err(); err != nil {
		collegeLogger.Error("Cursor error while fetching college applications by status", map[string]interface{}{
			"user_id": userID,
			"status":  status,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return applications, nil
}

func (r *mongoCollegeRepository) GetByPriority(ctx context.Context, userID string, priority string) ([]*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	filter := bson.M{
		"user_id":  objectID,
		"priority": priority,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		collegeLogger.Error("Failed to find college applications by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var applications []*models.CollegeApplication
	for cursor.Next(ctx) {
		var application models.CollegeApplication
		if err := cursor.Decode(&application); err != nil {
			collegeLogger.Error("Failed to decode college application", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		applications = append(applications, &application)
	}

	if err := cursor.Err(); err != nil {
		collegeLogger.Error("Cursor error while fetching college applications by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return applications, nil
}

func (r *mongoCollegeRepository) GetUpcomingDeadlines(ctx context.Context, userID string, days int) ([]*models.CollegeApplication, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		collegeLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"user_id": objectID,
		"deadline": bson.M{
			"$gte": now,
			"$lte": endDate,
		},
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "deadline", Value: 1}}) // Sort by deadline ascending

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		collegeLogger.Error("Failed to find college applications with upcoming deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var applications []*models.CollegeApplication
	for cursor.Next(ctx) {
		var application models.CollegeApplication
		if err := cursor.Decode(&application); err != nil {
			collegeLogger.Error("Failed to decode college application", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		applications = append(applications, &application)
	}

	if err := cursor.Err(); err != nil {
		collegeLogger.Error("Cursor error while fetching college applications with upcoming deadlines", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	collegeLogger.Info("Retrieved college applications with upcoming deadlines", map[string]interface{}{
		"user_id": userID,
		"days":    days,
		"count":   len(applications),
	})

	return applications, nil
}
