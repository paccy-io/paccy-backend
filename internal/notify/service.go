package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/notify/models"
	"github.com/paccy-io/paccy-backend/internal/notify/providers"
	"github.com/paccy-io/paccy-backend/internal/notify/repositories"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"github.com/paccy-io/paccy-backend/pkg/utils"
)

// BURADA KALDIN BASTAN DA YAZABILIRSIN

type NotifyService interface {
	CreateProjectQuota(ctx context.Context, quota *dto.QuotaDTO) error
	SendNotification(ctx context.Context, user *dto.SendNotificationDTO) error
	BlockUserQuota(ctx context.Context, userID string, quotaType models.QuotaType) error
	UnblockUserQuota(ctx context.Context, userID string, quotaType models.QuotaType) error
}

type notifyService struct {
	resendProvider   *providers.ResendProvider
	gmailProvider    *providers.GmailSmtpProvider
	userQuotaRepo    repositories.UserQuotaRepository
	projectQuotaRepo repositories.ProjectQuotaRepository
	validator        *validator.Validate
	log              *logger.Logger
}

func NewService(
	resendProvider *providers.ResendProvider,
	gmailProvider *providers.GmailSmtpProvider,
	validator *validator.Validate,
	userQuotaRepo repositories.UserQuotaRepository,
	projectQuotaRepo repositories.ProjectQuotaRepository,
) NotifyService {
	return &notifyService{
		resendProvider:   resendProvider,
		gmailProvider:    gmailProvider,
		userQuotaRepo:    userQuotaRepo,
		projectQuotaRepo: projectQuotaRepo,
		validator:        validator,
		log:              logger.Notify(),
	}
}

// CreateProjectQuota creates a new project quota with the default values
func (s *notifyService) CreateProjectQuota(ctx context.Context, quota *dto.QuotaDTO) error {
	s.log.WithContext(map[string]interface{}{
		"resender_daily":   quota.ResenderQuota.DailyQuota,
		"resender_monthly": quota.ResenderQuota.MonthlyQuota,
		"gmail_daily":      quota.GmailSmtpQuota.DailyQuota,
		"gmail_monthly":    quota.GmailSmtpQuota.MonthlyQuota,
	}).Info("Creating project quota")

	// Create a new project quota with default values
	err := s.projectQuotaRepo.CreateProjectQuota(ctx, quota)
	if err != nil {
		s.log.WithError(err).Error("Failed to create project quota")
		return fmt.Errorf("failed to create project quota: %w", err)
	}

	s.log.Info("Project quota created successfully")
	return nil
}

// SendNotification sends an email notification, checks quotas, and switches provider if needed
func (s *notifyService) SendNotification(ctx context.Context, user *dto.SendNotificationDTO) error {
	logCtx := s.log.WithUserID(user.UserID).WithField("template", user.Template).WithField("to", user.To)
	logCtx.Info("Sending notification")

	// 1. Check user quota (Notification)
	// 2. Check project quota (active provider)
	// 3. Send email via active provider
	// 4. Increase usage for both user and project

	// check if the template exists
	if !IsValidTemplateName(user.Template + ".html") {
		logCtx.Error("Invalid template name")
		return ErrInvalidTemplateName
	}

	data, err := RenderTemplate(user.Template+".html", map[string]interface{}{
		"Username": user.Username,
		"Link":     user.Link, // jwt token will be included in the link
		"Year":     time.Now().Year(),
	})

	if err != nil {
		logCtx.WithError(err).Error("Template rendering failed")
		return ErrTemplateRendering
	}
	logCtx.Debug("Template rendered successfully")

	// Convert userID to ObjectID
	userObjID, err := utils.ValidateObjectID(user.UserID, s.validator)
	if err != nil {
		logCtx.WithError(err).Error("Invalid user ID")
		return ErrInvalidInput
	}

	// Get active provider from project quota
	activeProvider, err := s.projectQuotaRepo.GetActiveProvider(ctx)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get active provider")
		return err
	}

	logCtx.WithField("provider", activeProvider).Debug("Active provider retrieved successfully")

	// Check user quota
	err = s.userQuotaRepo.IncreaseUserUsage(ctx, userObjID, string(models.QuotaTypeAuth))
	if err != nil {
		logCtx.WithError(err).Error("Failed to increase user quota usage")
		return err
	}
	logCtx.Debug("User quota usage increased successfully")

	// Send email via the selected provider
	logCtx.WithField("provider", activeProvider).Info("Sending email via provider")
	switch activeProvider {
	case models.ProviderTypeResender:
		err = s.resendProvider.SendEmail(ctx, user.To, user.Subject, data)
	case models.ProviderTypeGmailSmtp:
		err = s.gmailProvider.SendEmail(ctx, user.To, user.Subject, data)
	default:
		logCtx.WithField("provider", activeProvider).Error("Invalid provider type")
		return ErrInvalidProviderType
	}
	if err != nil {
		logCtx.WithError(err).Error("Failed to send email")
		
		// Send Discord notification for email failure
		logger.SendEmailFailureNotification(user.UserID, user.To, user.Subject, user.Template, string(activeProvider), err.Error())
		
		// decrease user quota usage if sending fails
		decreaseErr := s.userQuotaRepo.DecreaseUserUsage(ctx, userObjID, string(models.QuotaTypeNotification))
		if decreaseErr != nil {
			logCtx.WithError(decreaseErr).Error("Failed to decrease user quota usage after email failure")
			return decreaseErr
		}
		return err
	}
	logCtx.Info("Email sent successfully")

	// Send Discord notification for email success
	logger.SendEmailSuccessNotification(user.UserID, user.To, user.Subject, user.Template, string(activeProvider))

	// Increase project quota usage
	err = s.projectQuotaRepo.IncreaseUsage(ctx)
	if err != nil {
		// log the error but do not return it, as we already sent the email
		logCtx.WithError(err).Error("Failed to increase project quota usage after email sent")
		return err
	}

	return nil
}

func (s *notifyService) BlockUserQuota(ctx context.Context, userID string, quotaType models.QuotaType) error {
	logCtx := s.log.WithFields(map[string]interface{}{
		"userID":    userID,
		"quotaType": quotaType,
	})
	logCtx.Info("Blocking user quota")

	userObjID, err := utils.ValidateObjectID(userID, s.validator)
	if err != nil {
		logCtx.WithError(err).Error("Invalid user ID")
		return ErrInvalidInput
	}

	err = s.userQuotaRepo.BlockUserQuota(ctx, userObjID, string(quotaType))
	if err != nil {
		logCtx.WithError(err).Error("Failed to block user quota")
		return err
	}

	logCtx.Info("User quota blocked successfully")
	return nil
}

func (s *notifyService) UnblockUserQuota(ctx context.Context, userID string, quotaType models.QuotaType) error {
	logCtx := s.log.WithFields(map[string]interface{}{
		"userID":    userID,
		"quotaType": quotaType,
	})
	logCtx.Info("Unblocking user quota")

	userObjID, err := utils.ValidateObjectID(userID, s.validator)
	if err != nil {
		logCtx.WithError(err).Error("Invalid user ID")
		return ErrInvalidInput
	}

	err = s.userQuotaRepo.UnblockUserQuota(ctx, userObjID, string(quotaType))
	if err != nil {
		logCtx.WithError(err).Error("Failed to unblock user quota")
		return err
	}

	logCtx.Info("User quota unblocked successfully")
	return nil
}
