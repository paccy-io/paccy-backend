package user

import (
	"context"

	"log"
	"time"

	"github.com/google/uuid"
	"github.com/paccy-io/paccy-backend/internal/user/models"

	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Query operations
	GetUsers(ctx context.Context, offset, limit int) ([]models.User, int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]models.User, int, error)
	GetUsersByRole(ctx context.Context, role models.UserRole, offset, limit int) ([]models.User, int, error)
	GetActiveUsers(ctx context.Context, offset, limit int) ([]models.User, int, error)

	// Utility operations
	Exists(ctx context.Context, id primitive.ObjectID) bool
	ExistsByEmail(ctx context.Context, email string) bool
	ExistsByUsername(ctx context.Context, username string) bool
	Count(ctx context.Context) (int64, error)
	CountByRole(ctx context.Context, role models.UserRole) (int64, error)
}

type mongoUserRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
}

func NewMongoUserRepository(db *mongo.Database) UserRepository {
	collection := db.Collection(models.GetCollectionName())

	repo := &mongoUserRepository{
		collection: collection,
		database:   db,
	}

	repo.createIndexes()

	return repo
}

// decodeUsers is a helper function to decode cursor results
func decodeUsers(ctx context.Context, cursor *mongo.Cursor) ([]models.User, error) {
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *mongoUserRepository) createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("email_unique"),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("username_unique"),
		},
		{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uuid_unique"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("created_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "roles", Value: 1}},
			Options: options.Index().SetName("roles"),
		},
		{
			Keys:    bson.D{{Key: "verification_level", Value: 1}},
			Options: options.Index().SetName("verification_level"),
		},
		{
			Keys:    bson.D{{Key: "banned_at", Value: 1}},
			Options: options.Index().SetName("banned_at"),
		},
		{
			Keys:    bson.D{{Key: "auth_provider", Value: 1}},
			Options: options.Index().SetName("auth_provider"),
		},
		// Eğer phone unique olacaksa:
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true).SetName("phone_unique"),
		},
		{
			Keys:    bson.D{{Key: "deleted_at", Value: 1}},
			Options: options.Index().SetName("deleted_at"),
		},
		// Full-text arama indeksi (opsiyonel)
		{
			Keys: bson.D{
				{Key: "email", Value: "text"},
				{Key: "username", Value: "text"},
				{Key: "first_name", Value: "text"},
				{Key: "last_name", Value: "text"},
			},
			Options: options.Index().SetName("search_text"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		// Log the error but do not fail the application
		log.Printf("Failed to create indexes: %v", err)
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if r.ExistsByEmail(ctx, user.Email) || r.ExistsByUsername(ctx, user.Username) {
		return nil, ErrUserExists
	}
	user.BeforeCreate()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		log := logger.User()
		log.WithError(err).Error("Error inserting user")
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}

	return r.GetByID(ctx, user.ID)
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	filter := bson.M{"_id": id, "deleted_at": nil, "roles": bson.M{"$nin": []models.UserRole{models.RoleBanned}}} // Exclude banned users

	var u models.User
	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *mongoUserRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, error) {
	filter := bson.M{"uuid": uuid, "deleted_at": nil, "roles": bson.M{"$nin": []models.UserRole{models.RoleBanned}}} // Exclude banned users
	var u models.User
	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *mongoUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	filter := bson.M{"email": email, "deleted_at": nil, "roles": bson.M{"$nin": []models.UserRole{models.RoleBanned}}} // Exclude banned users
	var u models.User
	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *mongoUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	filter := bson.M{"username": username, "deleted_at": nil, "roles": bson.M{"$nin": []models.UserRole{models.RoleBanned}}} // Exclude banned users
	var u models.User
	err := r.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *mongoUserRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	filter := bson.M{"_id": id, "deleted_at": nil}

	delete(fields, "_id")

	update := bson.M{"$set": fields}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *mongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id, "deleted_at": nil}

	// Soft delete: set deleted_at to current time
	update := bson.M{"$set": bson.M{"deleted_at": time.Now().UTC(), "updated_at": time.Now().UTC()}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *mongoUserRepository) GetUsers(ctx context.Context, offset, limit int) ([]models.User, int, error) {
	filter := bson.M{"deleted_at": nil} // Do not exclude banned users here, as we want to return all users except those who are soft-deleted; there is a separate method named GetActiveUsers that excludes banned users
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	users, err := decodeUsers(ctx, cursor)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, int(count), nil
}

func (r *mongoUserRepository) Search(ctx context.Context, query string, offset, limit int) ([]models.User, int, error) {
	filter := bson.M{
		"$text":      bson.M{"$search": query},
		"deleted_at": nil,
		"roles":      bson.M{"$nin": []models.UserRole{models.RoleBanned}}, // Exclude banned users
	}
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	users, err := decodeUsers(ctx, cursor)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, int(count), nil
}

func (r *mongoUserRepository) GetUsersByRole(ctx context.Context, role models.UserRole, offset, limit int) ([]models.User, int, error) {
	// validate role in service
	filter := bson.M{"roles": role, "deleted_at": nil} // no need to exclude banned users since they are already set to RoleBanned and have no other roles
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	users, err := decodeUsers(ctx, cursor)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, int(count), nil
}

func (r *mongoUserRepository) GetActiveUsers(ctx context.Context, offset, limit int) ([]models.User, int, error) {
	filter := bson.M{"deleted_at": nil, "roles": bson.M{"$nin": []models.UserRole{models.RoleBanned}}} // Exclude banned users
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	users, err := decodeUsers(ctx, cursor)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, int(count), nil
}

func (r *mongoUserRepository) Exists(ctx context.Context, id primitive.ObjectID) bool {
	filter := bson.M{"_id": id, "deleted_at": nil} // we may want to check only active users, so we don't check roles here (banned users included)
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("Error checking existence of user with ID %s: %v", id.Hex(), err)
		return false
	}
	return count > 0
}

func (r *mongoUserRepository) ExistsByEmail(ctx context.Context, email string) bool {
	filter := bson.M{"email": email, "deleted_at": nil} // we don't exclude banned users here, as we want to check if the email exists regardless of the user's status
	count, err := r.collection.CountDocuments(ctx, filter)
	return err == nil && count > 0
}

func (r *mongoUserRepository) ExistsByUsername(ctx context.Context, username string) bool {
	filter := bson.M{"username": username, "deleted_at": nil}
	count, err := r.collection.CountDocuments(ctx, filter)
	return err == nil && count > 0
}

func (r *mongoUserRepository) Count(ctx context.Context) (int64, error) {
	filter := bson.M{"deleted_at": nil} // Count only active users, excluding soft-deleted users
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *mongoUserRepository) CountByRole(ctx context.Context, role models.UserRole) (int64, error) {
	if !models.IsValidRole(string(role)) {
		return 0, ErrInvalidRole
	}

	filter := bson.M{"roles": role, "deleted_at": nil} // Count only active users with the specified role, excluding soft-deleted users
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
