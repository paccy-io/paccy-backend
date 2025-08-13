package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/user/dto"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	. "github.com/paccy-io/paccy-backend/pkg/models"
	. "github.com/paccy-io/paccy-backend/pkg/utils"
)

type UserService interface {
	Create(ctx context.Context, user *dto.CreateUserDTO) (*models.SafeUser, error)
	GetByID(ctx context.Context, id string) (*models.SafeUser, error)
	GetByUUID(ctx context.Context, uuid string) (*models.SafeUser, error)
	GetByEmail(ctx context.Context, email string) (*models.SafeUser, error)
	GetByUsername(ctx context.Context, username string) (*models.SafeUser, error)
	UpdateUserVerificationLevel(ctx context.Context, id string, level models.VerificationLevel) error
	UpdateUserFields(ctx context.Context, id string, fields *dto.UpdateUserDTO) error
	UpdateProfile(ctx context.Context, id string, profile *dto.UpdateUserProfileDTO) error
	UpdatePreferences(ctx context.Context, id string, prefs *dto.UpdateUserPreferencesDTO) error
	//VerifyEmail(ctx context.Context, id string) error
	//VerifyPhone(ctx.Context, id string) error

	Delete(ctx context.Context, id string) error
	VerifyPassword(ctx context.Context, id string, password string) (models.SafeUser, bool, error)

	// Query operations
	GetUsers(ctx context.Context, paginationQuery *PaginationQuery) ([]models.SafeUser, int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]models.SafeUser, int, error)
	GetUsersByRole(ctx context.Context, role models.UserRole, offset, limit int) ([]models.SafeUser, int, error)
	GetActiveUsers(ctx context.Context, offset, limit int) ([]models.SafeUser, int, error)
	/*
		// Utility operations
		Exists(ctx context.Context, id string) bool
		ExistsByEmail(ctx context.Context, email string) bool
		ExistsByUsername(ctx context.Context, username string) bool
		Count(ctx context.Context) (int64, error)
		CountByRole(ctx context.Context, role models.UserRole) (int64, error)

		// admin operations
		BanUser(ctx context.Context, id string, reason string) error
		UnbanUser(ctx context.Context, id string) error
		SetRole(ctx context.Context, id string, role models.UserRole) error
		GetUserRoles(ctx context.Context, id string) ([]models.UserRole, error)
	*/
}

type userService struct {
	repo     UserRepository
	validate *validator.Validate
	log      *logger.Logger
}

func NewUserService(repo UserRepository, validate *validator.Validate) UserService {
	return &userService{
		repo:     repo,
		validate: validate,
		log:      logger.User(),
	}
}

func (s *userService) Create(ctx context.Context, user *dto.CreateUserDTO) (*models.SafeUser, error) {
	s.log.WithField("email", user.Email).WithField("username", user.Username).Info("Creating user")

	if err := s.validate.Struct(user); err != nil {
		s.log.WithError(err).WithField("email", user.Email).Error("User validation failed")
		return nil, ErrInvalidInput
	}

	password, err := HashPassword(user.Password)
	if err != nil {
		s.log.WithError(err).Error("Password hashing failed")
		return nil, ErrUnexpected
	}

	modelUser := &models.User{
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Password:      password,
		AuthProvider:  models.AuthProvider(user.AuthProvider),
		ProviderID:    user.ProviderID,
		TermsAccepted: user.TermsAccepted,
	}

	modelUser.BeforeCreate()

	createdUser, err := s.repo.Create(ctx, modelUser)
	if err != nil {
		s.log.WithError(err).WithField("email", user.Email).Error("Failed to create user in database")
		return nil, err
	}

	s.log.WithUserID(createdUser.ID.Hex()).WithField("email", user.Email).Info("User created successfully")

	return createdUser.ToSafe(), nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*models.SafeUser, error) {
	logCtx := s.log.WithField("userID", id)
	logCtx.Debug("Getting user by ID")

	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		logCtx.WithError(err).Error("Invalid user ID")
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user from database")
		return nil, err
	}

	if user.IsUserBanned() {
		logCtx.Warn("Attempted to get banned user")
		return nil, ErrUserBanned
	}

	logCtx.WithField("email", user.Email).Debug("User retrieved successfully")
	return user.ToSafe(), nil
}

func (s *userService) GetByUUID(ctx context.Context, uuid string) (*models.SafeUser, error) {
	uuidObj, err := ValidateUUID(uuid, s.validate)
	if err != nil {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByUUID(ctx, uuidObj)
	if err != nil {
		return nil, err
	}

	if user.IsUserBanned() {
		return nil, ErrUserBanned
	}

	return user.ToSafe(), nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*models.SafeUser, error) {
	logCtx := s.log.WithField("email", email)
	logCtx.Debug("Getting user by email")

	if err := ValidateEmail(email, s.validate); err != nil {
		logCtx.WithError(err).Error("Invalid email format")
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user by email")
		return nil, ErrUserNotFound
	}

	if user.IsUserBanned() {
		logCtx.WithUserID(user.ID.Hex()).Warn("Attempted to get banned user")
		return nil, ErrUserBanned
	}

	if user.VerificationLevel < models.EmailVerified {
		logCtx.WithUserID(user.ID.Hex()).Warn("Attempted to get unverified user")
		return nil, ErrUserNotVerified
	}

	logCtx.WithUserID(user.ID.Hex()).Debug("User retrieved successfully")
	return user.ToSafe(), nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*models.SafeUser, error) {
	sanitized_username, err := TrimAndValidateString(username, s.validate)
	if err != nil {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByUsername(ctx, sanitized_username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsUserBanned() {
		return nil, ErrUserBanned
	}

	if user.VerificationLevel < models.EmailVerified {
		return nil, ErrUserNotVerified
	}

	return user.ToSafe(), nil
}

// update user verification level
func (s *userService) UpdateUserVerificationLevel(ctx context.Context, id string, level models.VerificationLevel) error {
	logCtx := s.log.WithField("userID", id).WithField("verificationLevel", level)
	logCtx.Info("Updating user verification level")

	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		logCtx.WithError(err).Error("Invalid user ID")
		return ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user for verification level update")
		return ErrUserNotFound
	}

	oldLevel := user.VerificationLevel
	user.VerificationLevel = level
	user.BeforeUpdate()

	err = s.repo.UpdateFields(ctx, idObj, map[string]interface{}{
		"verification_level": level,
	})
	if err != nil {
		logCtx.WithError(err).Error("Failed to update verification level in database")
		return err
	}

	logCtx.WithField("oldLevel", oldLevel).WithField("newLevel", level).WithField("email", user.Email).Info("User verification level updated successfully")
	return nil
}

func (s *userService) UpdateUserFields(ctx context.Context, id string, fields *dto.UpdateUserDTO) error {
	if err := s.validate.Struct(fields); err != nil {
		return ErrInvalidInput
	}

	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		return ErrInvalidInput
	}

	if fields.Username != nil && *fields.Username != "" {
		existingUser, err := s.repo.GetByUsername(ctx, *fields.Username)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return ErrUnexpected
		}

		if existingUser != nil && existingUser.ID != idObj {
			return ErrUserExists
		}

		username := strings.ToLower(strings.TrimSpace(*fields.Username))
		fields.Username = &username
	}

	updateFields := make(map[string]interface{})
	if fields.Username != nil {
		updateFields["username"] = *fields.Username
	}
	if fields.FirstName != nil {
		updateFields["first_name"] = *fields.FirstName
	}
	if fields.LastName != nil {
		updateFields["last_name"] = *fields.LastName
	}

	// find user and call BeforeUpdate on the user model
	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		return ErrUserNotFound
	}
	user.BeforeUpdate()

	return s.repo.UpdateFields(ctx, idObj, updateFields)
}

func (s *userService) UpdateProfile(ctx context.Context, id string, profile *dto.UpdateUserProfileDTO) error {
	if err := s.validate.Struct(profile); err != nil {
		return ErrInvalidInput
	}

	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		return ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		return ErrUserNotFound
	}

	fields := make(map[string]interface{})
	if profile.AvatarURL != nil {
		fields["profile.avatar_url"] = *profile.AvatarURL
	}
	if profile.Bio != nil {
		fields["profile.bio"] = *profile.Bio
	}
	if profile.GraduationYear != nil {
		fields["profile.graduation_year"] = *profile.GraduationYear
	}
	if profile.School != nil {
		fields["profile.school"] = *profile.School
	}
	if profile.Website != nil {
		fields["profile.website"] = *profile.Website
	}
	if profile.SocialLinks != nil {
		fields["profile.social_links"] = profile.SocialLinks
	}
	if profile.Birthdate != nil {
		// convert month-day-year to time
		birthdate, err := time.Parse("01-02-2006", *profile.Birthdate)
		if err != nil {
			return ErrInvalidInput
		}
		fields["profile.birthdate"] = birthdate
	}
	if profile.Location != nil {
		fields["profile.location"] = *profile.Location
	}
	if profile.IsPublicProfile != nil {
		fields["profile.is_public_profile"] = *profile.IsPublicProfile
	}

	user.BeforeUpdate()

	return s.repo.UpdateFields(ctx, idObj, fields)
}

func (s *userService) UpdatePreferences(ctx context.Context, id string, prefs *dto.UpdateUserPreferencesDTO) error {
	if err := s.validate.Struct(prefs); err != nil {
		return ErrInvalidInput
	}

	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		return ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		return ErrUserNotFound
	}

	fields := make(map[string]interface{})
	if prefs.Theme != nil {
		fields["preferences.theme"] = *prefs.Theme
	}
	if prefs.Language != nil {
		fields["preferences.language"] = *prefs.Language
	}
	if prefs.Timezone != nil {
		fields["preferences.timezone"] = *prefs.Timezone
	}
	if prefs.EmailNotifications != nil {
		fields["preferences.email_notifications"] = *prefs.EmailNotifications
	}
	if prefs.PushNotifications != nil {
		fields["preferences.push_notifications"] = *prefs.PushNotifications
	}

	user.BeforeUpdate()

	return s.repo.UpdateFields(ctx, idObj, fields)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	idObj, err := ValidateObjectID(id, s.validate)
	if err != nil {
		return ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, idObj)
	if err != nil {
		return ErrUserNotFound
	}

	user.BeforeUpdate()

	return s.repo.Delete(ctx, idObj)
}

func (s *userService) GetUsers(ctx context.Context, paginationQuery *PaginationQuery) ([]models.SafeUser, int, error) {
	// validate pagination parameters with google validator
	if err := s.validate.Struct(paginationQuery); err != nil {
		return nil, 0, ErrInvalidInput
	}
	offset := paginationQuery.GetOffset()
	limit := paginationQuery.Limit

	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidInput
	}

	users, total, err := s.repo.GetUsers(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	safeUsers := make([]models.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = *user.ToSafe()
	}

	return safeUsers, total, nil
}

func (s *userService) Search(ctx context.Context, query string, offset, limit int) ([]models.SafeUser, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidInput
	}

	if len(query) < 3 {
		return nil, 0, ErrInvalidQuery
	}

	users, total, err := s.repo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	safeUsers := make([]models.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = *user.ToSafe()
	}

	return safeUsers, total, nil
}

func (s *userService) GetUsersByRole(ctx context.Context, role models.UserRole, offset, limit int) ([]models.SafeUser, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidInput
	}

	users, total, err := s.repo.GetUsersByRole(ctx, role, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	safeUsers := make([]models.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = *user.ToSafe()
	}

	return safeUsers, total, nil
}

func (s *userService) GetActiveUsers(ctx context.Context, offset, limit int) ([]models.SafeUser, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidInput
	}

	users, total, err := s.repo.GetActiveUsers(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	safeUsers := make([]models.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = *user.ToSafe()
	}

	return safeUsers, total, nil
}

func (s *userService) VerifyPassword(ctx context.Context, email, password string) (models.SafeUser, bool, error) {
	logCtx := s.log.WithField("email", email)
	logCtx.Debug("Verifying user password")

	// validate email
	if err := s.validate.Var(email, "required,email"); err != nil {
		logCtx.WithError(err).Error("Invalid email format for password verification")
		return models.SafeUser{}, false, ErrInvalidInput
	}

	// validate password
	if err := s.validate.Var(password, "required,min=8"); err != nil {
		logCtx.WithError(err).Error("Invalid password format for verification")
		return models.SafeUser{}, false, ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get user for password verification")
		return models.SafeUser{}, false, ErrUserNotFound
	}

	verified := VerifyPassword(user.Password, password)

	if !verified {
		logCtx.WithUserID(user.ID.Hex()).Warn("Password verification failed")
		return models.SafeUser{}, false, nil
	}

	if user.VerificationLevel < models.EmailVerified {
		logCtx.WithUserID(user.ID.Hex()).Warn("Password verified but user not email verified")
		return models.SafeUser{}, false, ErrUserNotVerified
	}

	logCtx.WithUserID(user.ID.Hex()).Info("Password verified successfully")
	return *user.ToSafe(), true, nil
}
