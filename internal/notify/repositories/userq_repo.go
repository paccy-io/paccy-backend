package repositories

import (
	"context"
	"fmt"

	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/notify/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserQuotaRepository interface {
	CreateUserQuota(ctx context.Context, userID primitive.ObjectID) error
	GetUserQuota(ctx context.Context, userID primitive.ObjectID) (*models.UserQuota, error)
	IncreaseUserUsage(ctx context.Context, userID primitive.ObjectID, quotaType string) error // if user has no quota, it will be created with default values and then usage will be increased
	DecreaseUserUsage(ctx context.Context, userID primitive.ObjectID, quotaType string) error
	BlockUserQuota(ctx context.Context, userID primitive.ObjectID, quotaType string) error
	UnblockUserQuota(ctx context.Context, userID primitive.ObjectID, quotaType string) error
}

type mongoUserQuotaRepository struct {
	collection *mongo.Collection
	database   *mongo.Database
	quota      *dto.UserQuotaDTO
}

func NewMongoUserQuotaRepository(db *mongo.Database, quota *dto.UserQuotaDTO) UserQuotaRepository {
	collection := db.Collection(models.GetUserQuotaCollectionName())
	return &mongoUserQuotaRepository{
		collection: collection,
		database:   db,
		quota:      quota,
	}
}

func (r *mongoUserQuotaRepository) CreateUserQuota(ctx context.Context, userID primitive.ObjectID) error {
	// Check if the user quota already exists
	existingUserQuota, err := r.GetUserQuota(ctx, userID)
	if err != nil && err != ErrNoUserQuotaFound {
		return ErrUnexpected
	}

	// If the user quota already exists, return an error
	if existingUserQuota != nil {
		return ErrUserQuotaExists
	}

	model := &models.UserQuota{
		UserID: userID,
	}

	model.BeforeCreate(r.quota)

	_, err = r.collection.InsertOne(ctx, model)
	if err != nil {
		return ErrUnexpected
	}
	return nil
}

func (r *mongoUserQuotaRepository) GetUserQuota(ctx context.Context, userID primitive.ObjectID) (*models.UserQuota, error) {
	var userQuota models.UserQuota
	filter := map[string]interface{}{"user_id": userID}

	err := r.collection.FindOne(ctx, filter).Decode(&userQuota)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoUserQuotaFound
		}
		return nil, ErrUnexpected
	}

	return &userQuota, nil
}

func (r *mongoUserQuotaRepository) IncreaseUserUsage(ctx context.Context, userID primitive.ObjectID, quotaType string) error {
	userQuota, err := r.GetUserQuota(ctx, userID)
	if err != nil {
		if err == ErrNoUserQuotaFound {
			// Create a new user quota with default values if it doesn't exist
			err = r.CreateUserQuota(ctx, userID)
			if err != nil {
				fmt.Println("Error creating user quota INLINE!:", err)
				return ErrUnexpected
			}
			userQuota, _ = r.GetUserQuota(ctx, userID) // re-fetch the newly created quota
		} else {
			fmt.Println("Error getting user quota OUTLINE!:", err)
			return ErrUnexpected
		}
	}

	var baseQuota *models.BaseQuota

	switch quotaType {
	case string(models.QuotaTypeAuth):
		baseQuota = &userQuota.Auth
	case string(models.QuotaTypeSecurity):
		baseQuota = &userQuota.Security
	case string(models.QuotaTypeNotification):
		baseQuota = &userQuota.Notification
	default:
		return ErrInvalidProviderType
	}

	boolean := baseQuota.IncrementUsage()

	if !boolean {
		fmt.Println("Error incrementing usage for quota type:", quotaType)
		return ErrUnexpected
	}

	userQuota.BeforeUpdate()

	_, err = r.collection.UpdateOne(ctx, map[string]interface{}{"user_id": userID}, map[string]interface{}{"$set": userQuota})
	if err != nil {
		return ErrUnexpected
	}

	return nil
}

func (r *mongoUserQuotaRepository) DecreaseUserUsage(ctx context.Context, userID primitive.ObjectID, quotaType string) error {
	userQuota, err := r.GetUserQuota(ctx, userID)
	if err != nil {
		return ErrUnexpected
	}

	var baseQuota *models.BaseQuota

	switch quotaType {
	case string(models.QuotaTypeAuth):
		baseQuota = &userQuota.Auth
	case string(models.QuotaTypeSecurity):
		baseQuota = &userQuota.Security
	case string(models.QuotaTypeNotification):
		baseQuota = &userQuota.Notification
	default:
		return ErrInvalidProviderType
	}

	boolean := baseQuota.DecrementUsage()

	if !boolean {
		return ErrUnexpected
	}

	userQuota.BeforeUpdate()

	_, err = r.collection.UpdateOne(ctx, map[string]interface{}{"user_id": userID}, map[string]interface{}{"$set": userQuota})
	if err != nil {
		return ErrUnexpected
	}

	return nil
}

func (r *mongoUserQuotaRepository) BlockUserQuota(ctx context.Context, userID primitive.ObjectID, quotaType string) error {
	userQuota, err := r.GetUserQuota(ctx, userID)
	if err != nil {
		return ErrUnexpected
	}

	var baseQuota *models.BaseQuota

	switch quotaType {
	case string(models.QuotaTypeAuth):
		baseQuota = &userQuota.Auth
	case string(models.QuotaTypeSecurity):
		baseQuota = &userQuota.Security
	case string(models.QuotaTypeNotification):
		baseQuota = &userQuota.Notification
	default:
		return ErrInvalidProviderType
	}

	baseQuota.Block()

	userQuota.BeforeUpdate()

	_, err = r.collection.UpdateOne(ctx, map[string]interface{}{"user_id": userID}, map[string]interface{}{"$set": userQuota})
	if err != nil {
		return ErrUnexpected
	}

	return nil
}

func (r *mongoUserQuotaRepository) UnblockUserQuota(ctx context.Context, userID primitive.ObjectID, quotaType string) error {
	userQuota, err := r.GetUserQuota(ctx, userID)
	if err != nil {
		return ErrUnexpected
	}

	var baseQuota *models.BaseQuota

	switch quotaType {
	case string(models.QuotaTypeAuth):
		baseQuota = &userQuota.Auth
	case string(models.QuotaTypeSecurity):
		baseQuota = &userQuota.Security
	case string(models.QuotaTypeNotification):
		baseQuota = &userQuota.Notification
	default:
		return ErrInvalidProviderType
	}

	baseQuota.Unblock()

	userQuota.BeforeUpdate()

	_, err = r.collection.UpdateOne(ctx, map[string]interface{}{"user_id": userID}, map[string]interface{}{"$set": userQuota})
	if err != nil {
		return ErrUnexpected
	}

	return nil
}
