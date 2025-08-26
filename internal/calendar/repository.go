package calendar

import (
	"context"
	"time"

	"github.com/paccy-io/paccy-backend/internal/calendar/models"
	"github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var calendarLogger = logger.New("calendar")

type CalendarRepository interface {
	Create(ctx context.Context, event *models.Event) (*models.Event, error)
	GetByID(ctx context.Context, id string) (*models.Event, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Event, error)
	GetInboxEvents(ctx context.Context, userID string) ([]*models.Event, error)
	GetEventsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.Event, error)
	GetEventsByType(ctx context.Context, userID string, eventType string) ([]*models.Event, error)
	GetEventsByPriority(ctx context.Context, userID string, priority string) ([]*models.Event, error)
	GetUpcomingEvents(ctx context.Context, userID string, days int) ([]*models.Event, error)
	GetEventsByCollegeApplication(ctx context.Context, userID, collegeApplicationID string) ([]*models.Event, error)
	Update(ctx context.Context, id string, updates bson.M) (*models.Event, error)
	UpdateWithOperations(ctx context.Context, id string, setUpdates, unsetUpdates map[string]interface{}) (*models.Event, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, userID string) (int64, error)
}

type mongoCalendarRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
}

func NewMongoCalendarRepository(db *mongo.Database) CalendarRepository {
	collection := db.Collection(models.GetCollectionName())
	return &mongoCalendarRepository{
		collection: collection,
		database:   db,
	}
}

func (r *mongoCalendarRepository) Create(ctx context.Context, event *models.Event) (*models.Event, error) {
	calendarLogger.Info("Creating new event", map[string]interface{}{
		"title":   event.Title,
		"user_id": event.UserID.Hex(),
		"type":    string(event.Type),
	})

	event.BeforeCreate()

	result, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		calendarLogger.Error("Failed to insert event", map[string]interface{}{
			"error":   err.Error(),
			"title":   event.Title,
			"user_id": event.UserID.Hex(),
		})
		return nil, errors.ErrCreateFailed
	}

	event.ID = result.InsertedID.(primitive.ObjectID)
	calendarLogger.Info("Event created successfully", map[string]interface{}{
		"event_id": event.ID.Hex(),
		"title":    event.Title,
		"user_id":  event.UserID.Hex(),
	})

	return event, nil
}

func (r *mongoCalendarRepository) GetByID(ctx context.Context, id string) (*models.Event, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		calendarLogger.Error("Invalid ObjectID format", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	var event models.Event
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			calendarLogger.Info("Event not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		calendarLogger.Error("Failed to get event", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return &event, nil
}

func (r *mongoCalendarRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": userObjectID,
	}, options.Find().SetSort(bson.D{{Key: "date", Value: 1}, {Key: "created_at", Value: -1}}))
	if err != nil {
		calendarLogger.Error("Failed to find events for user", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	calendarLogger.Info("Retrieved events for user", map[string]interface{}{
		"user_id": userID,
		"count":   len(events),
	})

	return events, nil
}

func (r *mongoCalendarRepository) GetInboxEvents(ctx context.Context, userID string) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": userObjectID,
		"date":    nil, // Events without dates go to inbox
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		calendarLogger.Error("Failed to find inbox events for user", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode inbox event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching inbox events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return events, nil
}

func (r *mongoCalendarRepository) GetEventsByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	filter := bson.M{
		"user_id": userObjectID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
	if err != nil {
		calendarLogger.Error("Failed to find events by date range", map[string]interface{}{
			"user_id":    userID,
			"start_date": startDate,
			"end_date":   endDate,
			"error":      err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching events by date range", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return events, nil
}

func (r *mongoCalendarRepository) GetEventsByType(ctx context.Context, userID string, eventType string) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": userObjectID,
		"type":    eventType,
	}, options.Find().SetSort(bson.D{{Key: "date", Value: 1}, {Key: "created_at", Value: -1}}))
	if err != nil {
		calendarLogger.Error("Failed to find events by type", map[string]interface{}{
			"user_id":    userID,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching events by type", map[string]interface{}{
			"user_id":    userID,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return events, nil
}

func (r *mongoCalendarRepository) GetEventsByPriority(ctx context.Context, userID string, priority string) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":  userObjectID,
		"priority": priority,
	}, options.Find().SetSort(bson.D{{Key: "date", Value: 1}, {Key: "created_at", Value: -1}}))
	if err != nil {
		calendarLogger.Error("Failed to find events by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching events by priority", map[string]interface{}{
			"user_id":  userID,
			"priority": priority,
			"error":    err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	return events, nil
}

func (r *mongoCalendarRepository) GetUpcomingEvents(ctx context.Context, userID string, days int) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	filter := bson.M{
		"user_id": userObjectID,
		"date": bson.M{
			"$gte": now,
			"$lte": endDate,
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
	if err != nil {
		calendarLogger.Error("Failed to find upcoming events", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching upcoming events", map[string]interface{}{
			"user_id": userID,
			"days":    days,
			"error":   err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	calendarLogger.Info("Retrieved upcoming events", map[string]interface{}{
		"user_id": userID,
		"days":    days,
		"count":   len(events),
	})

	return events, nil
}

func (r *mongoCalendarRepository) Update(ctx context.Context, id string, updates bson.M) (*models.Event, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		calendarLogger.Error("Invalid ObjectID format for update", map[string]interface{}{
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

	var event models.Event
	if err := result.Decode(&event); err != nil {
		if err == mongo.ErrNoDocuments {
			calendarLogger.Warn("Event not found for update", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		calendarLogger.Error("Failed to update event", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	calendarLogger.Info("Event updated successfully", map[string]interface{}{
		"event_id": id,
		"title":    event.Title,
	})

	return &event, nil
}

func (r *mongoCalendarRepository) UpdateWithOperations(ctx context.Context, id string, setUpdates, unsetUpdates map[string]interface{}) (*models.Event, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		calendarLogger.Error("Invalid ObjectID format for update with operations", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	// Build update operations
	updateOps := bson.M{}
	
	// Add $set operations if there are any
	if len(setUpdates) > 0 {
		// Add updated_at timestamp
		setUpdates["updated_at"] = time.Now()
		updateOps["$set"] = setUpdates
	}
	
	// Add $unset operations if there are any
	if len(unsetUpdates) > 0 {
		updateOps["$unset"] = unsetUpdates
	}

	// If no operations, just update the timestamp
	if len(updateOps) == 0 {
		updateOps["$set"] = bson.M{"updated_at": time.Now()}
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		updateOps,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var event models.Event
	if err := result.Decode(&event); err != nil {
		if err == mongo.ErrNoDocuments {
			calendarLogger.Warn("Event not found for update with operations", map[string]interface{}{
				"id": id,
			})
			return nil, errors.ErrGetFailed
		}
		calendarLogger.Error("Failed to update event with operations", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return nil, errors.ErrUpdateFailed
	}

	calendarLogger.Info("Event updated successfully with operations", map[string]interface{}{
		"event_id": id,
		"title":    event.Title,
	})

	return &event, nil
}

func (r *mongoCalendarRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		calendarLogger.Error("Invalid ObjectID format for delete", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrInvalidObjectID
	}

	// Hard delete: permanently remove the document
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		calendarLogger.Error("Failed to delete event", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return errors.ErrDeleteFailed
	}

	if result.DeletedCount == 0 {
		calendarLogger.Warn("Event not found for deletion", map[string]interface{}{
			"id": id,
		})
		return errors.ErrGetFailed
	}

	calendarLogger.Info("Event deleted successfully", map[string]interface{}{
		"event_id": id,
	})

	return nil
}

func (r *mongoCalendarRepository) GetEventsByCollegeApplication(ctx context.Context, userID, collegeApplicationID string) ([]*models.Event, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	collegeAppObjectID, err := primitive.ObjectIDFromHex(collegeApplicationID)
	if err != nil {
		calendarLogger.Error("Invalid college application ObjectID format", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrInvalidObjectID
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":    userObjectID,
		"college_id": collegeAppObjectID,
	}, options.Find().SetSort(bson.D{{Key: "date", Value: 1}, {Key: "created_at", Value: -1}}))
	if err != nil {
		calendarLogger.Error("Failed to find events by college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrGetFailed
	}
	defer cursor.Close(ctx)

	var events []*models.Event
	for cursor.Next(ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			calendarLogger.Error("Failed to decode event", map[string]interface{}{
				"user_id":                userID,
				"college_application_id": collegeApplicationID,
				"error":                  err.Error(),
			})
			continue
		}
		events = append(events, &event)
	}

	if err := cursor.Err(); err != nil {
		calendarLogger.Error("Cursor error while fetching events by college application", map[string]interface{}{
			"user_id":                userID,
			"college_application_id": collegeApplicationID,
			"error":                  err.Error(),
		})
		return nil, errors.ErrGetFailed
	}

	calendarLogger.Info("Retrieved events by college application", map[string]interface{}{
		"user_id":                userID,
		"college_application_id": collegeApplicationID,
		"count":                  len(events),
	})

	return events, nil
}

func (r *mongoCalendarRepository) Count(ctx context.Context, userID string) (int64, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		calendarLogger.Error("Invalid user ObjectID format for count", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrInvalidObjectID
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userObjectID})
	if err != nil {
		calendarLogger.Error("Failed to count events", map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		})
		return 0, errors.ErrGetFailed
	}

	return count, nil
}
