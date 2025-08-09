package repositories

import (
	"context"
	"fmt"

	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/notify/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectQuotaRepository interface {
	CreateProjectQuota(ctx context.Context, quota *dto.QuotaDTO) error
	GetProjectQuota(ctx context.Context) (*models.ProjectQuota, error)
	GetActiveProvider(ctx context.Context) (models.ProviderType, error)
	GetActiveQuota(ctx context.Context) (*models.BaseProviderQuota, error)
	GetQuotaByProvider(ctx context.Context, providerType string) (*models.BaseProviderQuota, error)
	IncreaseUsage(ctx context.Context) error // reset usage will be handled here
}

type mongoProjectQuotaRepository struct {
	collection  *mongo.Collection
	database    *mongo.Database
	projectName string
}

func NewMongoProjectQuotaRepository(db *mongo.Database, projectName string) ProjectQuotaRepository {
	collection := db.Collection(models.GetCollectionName())
	return &mongoProjectQuotaRepository{
		collection:  collection,
		database:    db,
		projectName: projectName,
	}
}

func (r *mongoProjectQuotaRepository) CreateProjectQuota(ctx context.Context, quota *dto.QuotaDTO) error {
	// check if the project quota already exists
	existingProjectQuota, err := r.GetProjectQuota(ctx)
	if err != nil && err != ErrNoProjectQuotaFound {
		fmt.Println("Error checking existing project quota:", err)
		return ErrUnexpected
	}

	// if the project quota already exists, return an error
	if existingProjectQuota != nil {
		return ErrProjectQuotaExists
	}
	fmt.Println("Creating new project quota with default values")

	model := &models.ProjectQuota{
		ProjectName: r.projectName,
	}

	model.BeforeCreate(quota.ResenderQuota.DailyQuota, quota.ResenderQuota.MonthlyQuota, quota.GmailSmtpQuota.DailyQuota, quota.GmailSmtpQuota.MonthlyQuota)

	// create a new project quota
	_, err = r.collection.InsertOne(ctx, model)

	if err != nil {
		fmt.Println("Error inserting new project quota:", err)
		return ErrUnexpected
	}

	return nil
}

func (r *mongoProjectQuotaRepository) GetProjectQuota(ctx context.Context) (*models.ProjectQuota, error) {
	var quota models.ProjectQuota
	err := r.collection.FindOne(ctx, bson.M{"project_name": r.projectName}).Decode(&quota)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoProjectQuotaFound
		}
		return nil, ErrUnexpected
	}

	return &quota, nil
}

func (r *mongoProjectQuotaRepository) GetActiveProvider(ctx context.Context) (models.ProviderType, error) {
	// find the active provider
	quota, err := r.GetProjectQuota(ctx)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.ProviderType(""), ErrNoProjectQuotaFound
		}
		return models.ProviderType(""), ErrUnexpected
	}

	return quota.GetActiveProvider(), nil
}

func (r *mongoProjectQuotaRepository) GetActiveQuota(ctx context.Context) (*models.BaseProviderQuota, error) {
	projectQuota, err := r.GetProjectQuota(ctx)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoProjectQuotaFound
		}
		return nil, ErrUnexpected
	}

	activeProvider := projectQuota.GetActiveProvider()
	if activeProvider == models.ProviderType("") {
		return nil, ErrNoActiveProvider
	}
	var activeQuota *models.BaseProviderQuota

	switch activeProvider {
case projectQuota.Resender.ProviderType:
		activeQuota = &projectQuota.Resender
	case projectQuota.GmailSmtp.ProviderType:
		activeQuota = &projectQuota.GmailSmtp
	}

	return activeQuota, nil
}

func (r *mongoProjectQuotaRepository) GetQuotaByProvider(ctx context.Context, providerType string) (*models.BaseProviderQuota, error) {
	quota, err := r.GetProjectQuota(ctx)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoProjectQuotaFound
		}
		return nil, ErrUnexpected
	}

	if providerType == string(models.ProviderTypeResender) {
		return &quota.Resender, nil
	} else if providerType == string(models.ProviderTypeGmailSmtp) {
		return &quota.GmailSmtp, nil
	}

	return nil, ErrInvalidProviderType
}

func (r *mongoProjectQuotaRepository) IncreaseUsage(ctx context.Context) error {
	quota, err := r.GetProjectQuota(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNoProjectQuotaFound
		}
		return ErrUnexpected
	}

	providerQuota := quota.SwitchProvider()

	if providerQuota == models.ProviderType("") {
		return ErrNoActiveProvider
	}
	switch providerQuota {
case models.ProviderTypeResender:
		quota.Resender.IncreaseUsage()
	case models.ProviderTypeGmailSmtp:
		quota.GmailSmtp.IncreaseUsage()
	}

	// call BeforeUpdate to update the timestamp
	quota.BeforeUpdate()

	_, err = r.collection.UpdateOne(ctx, bson.M{"project_name": r.projectName}, bson.M{"$set": quota})

	if err != nil {
		return ErrUnexpected
	}
	return nil
}
