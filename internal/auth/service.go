package auth

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/notify"
	notifydto "github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/user"
	"github.com/paccy-io/paccy-backend/internal/user/dto"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"github.com/paccy-io/paccy-backend/pkg/utils"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)                         // returns JWT token or error
	RegisterAndSendVerification(ctx context.Context, dtoUser *dto.CreateUserDTO) (bool, error) // returns user creation status and error
	GenerateEmailVerificationToken(ctx context.Context, useruuid string) (string, error)       // generates email verification token
	SendEmailVerification(ctx context.Context, userID string) error                            // sends email verification notification
	VerifyEmail(ctx context.Context, token string) error                                       // verifies email using a token
	VerifyToken(ctx context.Context, token string) (*jwt.Claims, error)                        // verifies JWT token and returns claims
	//VerifyEmail(ctx context.Context, token string) error                              // verifies email using a token
	//TODO: ResetPassword(ctx context.Context, email string) (string, error)                   // sends reset password email and returns token
	//ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error // changes password for the user
}

type authService struct {
	validate      *validator.Validate
	jwtProvider   jwt.Provider
	userService   user.UserService
	notifyService notify.NotifyService
	log           *logger.Logger
}

func NewAuthService(validate *validator.Validate, jwtProvider jwt.Provider, userService user.UserService, notifyService notify.NotifyService) AuthService {
	return &authService{
		validate:      validate,
		jwtProvider:   jwtProvider,
		userService:   userService,
		notifyService: notifyService,
		log:           logger.Auth(),
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	s.log.WithField("email", email).Info("Login attempt")

	user, verified, err := s.userService.VerifyPassword(ctx, email, password)

	if err != nil {
		s.log.WithError(err).WithField("email", email).Error("Login failed - user verification error")
		return "", err
	}
	if !verified {
		s.log.WithField("email", email).Warn("Login failed - invalid credentials")
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtProvider.GenerateToken(user.ID, user.Username, user.Roles, user.VerificationLevel)
	if err != nil {
		s.log.WithError(err).WithUserID(user.ID.Hex()).Error("Login failed - token generation error")
		return "", err
	}

	s.log.WithUserID(user.ID.Hex()).WithField("email", email).Info("Login successful")
	return token, nil
}

// register user and send verification email
func (s *authService) RegisterAndSendVerification(ctx context.Context, dtoUser *dto.CreateUserDTO) (bool, error) {
	s.log.WithField("email", dtoUser.Email).Info("User registration attempt")

	user, err := s.userService.Create(ctx, dtoUser)
	bool := false // user creation status
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrUserExists) {
			return bool, err
		}
		s.log.WithError(err).Error("User registration failed")
		return bool, err
	}
	bool = true

	s.log.WithUserID(user.ID.Hex()).WithField("email", dtoUser.Email).Info("User created successfully, sending verification email")

	// Send Discord notification for new user registration
	logger.SendUserRegistrationNotification(user.ID.Hex(), user.Username, user.Email)

	err = s.SendEmailVerification(ctx, user.ID.Hex())
	if err != nil {
		s.log.WithError(err).WithUserID(user.ID.Hex()).Error("Failed to send verification email")
		return bool, err
	}

	return bool, nil
}

// Verify email using a token
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	s.log.WithField("token_length", len(token)).Info("Email verification attempt")

	claims, err := s.jwtProvider.VerifyToken(token)
	if err != nil {
		s.log.WithError(err).Error("Email verification failed - invalid token")
		return err
	}

	user, err := s.userService.GetByID(ctx, claims.ID.Hex())
	if err != nil {
		s.log.WithError(err).WithUserID(claims.ID.Hex()).Error("Email verification failed - user not found")
		return ErrUserNotFound
	}

	if claims.VerificationLevel < models.EmailVerified {
		s.log.WithUserID(user.ID.Hex()).Warn("Email verification failed - user not verified")
		return ErrUserNotVerified
	}

	if user.VerificationLevel >= models.EmailVerified {
		s.log.WithUserID(user.ID.Hex()).Info("Email already verified")
		return nil // already verified
	}

	err = s.userService.UpdateUserVerificationLevel(ctx, user.ID.Hex(), models.EmailVerified)
	if err != nil {
		s.log.WithError(err).WithUserID(user.ID.Hex()).Error("Failed to update user verification level")
		return err
	}

	s.log.WithUserID(user.ID.Hex()).Info("Email verified successfully")

	// Send Discord notification for email verification
	logger.SendEmailVerificationNotification(user.ID.Hex(), user.Email)

	return nil
}

// mail verification token
func (s *authService) GenerateEmailVerificationToken(ctx context.Context, userID string) (string, error) {
	s.log.WithUserID(userID).Info("Generating email verification token")

	// Generate a JWT token for email verification
	// check userID
	userIDObj, err := utils.ValidateObjectID(userID, s.validate)

	if err != nil {
		s.log.WithError(err).WithUserID(userID).Error("Invalid user ID for token generation")
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtProvider.GenerateEmailVerificationToken(userIDObj)
	if err != nil {
		s.log.WithError(err).WithUserID(userID).Error("Failed to generate email verification token")
		return "", err
	}

	s.log.WithUserID(userID).Info("Email verification token generated successfully")
	return token, nil
}

// send email verification token
func (s *authService) SendEmailVerification(ctx context.Context, userID string) error {
	s.log.WithUserID(userID).Info("Sending email verification")

	token, err := s.GenerateEmailVerificationToken(ctx, userID)
	if err != nil {
		s.log.WithError(err).WithUserID(userID).Error("Failed to generate verification token for email sending")
		return err
	}

	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		s.log.WithError(err).WithUserID(userID).Error("Failed to retrieve user for email verification")
		return err
	}

	s.log.WithUserID(userID).WithField("email", user.Email).Info("Sending verification email")

	err = s.notifyService.SendNotification(ctx, &notifydto.SendNotificationDTO{
		UserID:   user.ID.Hex(), //!!! Hex()
		Username: user.Username,
		To:       user.Email,
		Subject:  "Verify your email",
		Template: "verify_email",
		Link:     utils.GetEmailVerificationLink(token),
	})
	if err != nil {
		s.log.WithError(err).WithUserID(userID).WithField("email", user.Email).Error("Failed to send verification email")
		return err
	}

	s.log.WithUserID(userID).WithField("email", user.Email).Info("Verification email sent successfully")
	return nil
}

// verify jwt token
func (s *authService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	s.log.WithField("token_length", len(token)).Debug("Verifying JWT token")

	claims, err := s.jwtProvider.VerifyToken(token)
	if err != nil {
		s.log.WithError(err).Error("JWT token verification failed")
		return nil, err
	}

	s.log.WithUserID(claims.ID.Hex()).Debug("JWT token verified successfully")
	return claims, nil
}
