package auth

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/notify"
	notifydto "github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/user"
	"github.com/paccy-io/paccy-backend/internal/user/dto"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
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
}

func NewAuthService(validate *validator.Validate, jwtProvider jwt.Provider, userService user.UserService, notifyService notify.NotifyService) AuthService {
	return &authService{
		validate:      validate,
		jwtProvider:   jwtProvider,
		userService:   userService,
		notifyService: notifyService,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, verified, err := s.userService.VerifyPassword(ctx, email, password)

	if err != nil {
		return "", err
	}
	if !verified {
		return "", ErrInvalidCredentials
	}
	//!Test
	fmt.Println("id:", user.ID)


	token, err := s.jwtProvider.GenerateToken(user.ID, user.Username, user.Roles, user.VerificationLevel)
	if err != nil {
		return "", err
	}

	return token, nil
}

// register user and send verification email
func (s *authService) RegisterAndSendVerification(ctx context.Context, dtoUser *dto.CreateUserDTO) (bool, error) {
	user, err := s.userService.Create(ctx, dtoUser)
	bool := false // user creation status
	if err != nil {
		return bool, err
	}
	fmt.Println("User created successfully:", user)
	bool = true
	err = s.SendEmailVerification(ctx, user.ID.Hex())
	if err != nil {
		fmt.Println("Error sending email verification:", err)
		return bool, err
	}

	return bool, nil
}

// Verify email using a token
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	claims, err := s.jwtProvider.VerifyToken(token)
	if err != nil {
		return err
	}

	user, err := s.userService.GetByID(ctx, claims.ID.Hex())
	if err != nil {
		return ErrUserNotFound
	}

	if claims.VerificationLevel < models.EmailVerified {
		return ErrUserNotVerified
	}

	if user.VerificationLevel >= models.EmailVerified {
		return nil // already verified
	}

	err = s.userService.UpdateUserVerificationLevel(ctx, user.ID.Hex(), models.EmailVerified)
	if err != nil {
		return err
	}

	return nil
}

// mail verification token
func (s *authService) GenerateEmailVerificationToken(ctx context.Context, userID string) (string, error) {
	// Generate a JWT token for email verification

	// check userID
	userIDObj, err := utils.ValidateObjectID(userID, s.validate)

	if err != nil {
		fmt.Println(userID, userIDObj)
		fmt.Println("Error validating user ID:", err)
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtProvider.GenerateEmailVerificationToken(userIDObj)
	if err != nil {
		return "", err
	}
	return token, nil
}

// send email verification token
func (s *authService) SendEmailVerification(ctx context.Context, userID string) error {
	fmt.Println("this Sending email verification for user ID:", userID)
	token, err := s.GenerateEmailVerificationToken(ctx, userID)
	if err != nil {
		return err
	}
	fmt.Println("Generated email verification token:", token)
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	fmt.Println("User retrieved successfully:", user)
	err = s.notifyService.SendNotification(ctx, &notifydto.SendNotificationDTO{
		UserID:   user.ID.Hex(), //!!! Hex()
		Username: user.Username,
		To:       user.Email,
		Subject:  "Verify your email",
		Template: "verify_email",
		Link:     utils.GetEmailVerificationLink(token),
	})
	if err != nil {
		fmt.Println("11Error sending email verification:", err)
		return err
	}

	return nil
}

// verify jwt token
func (s *authService) VerifyToken(ctx context.Context, token string) (*jwt.Claims, error) {
	claims, err := s.jwtProvider.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
