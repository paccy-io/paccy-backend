package models

import (
	"strings"
	"time"

	"github.com/paccy-io/paccy-backend/pkg/models"
)

type User struct {
	models.BaseModel `bson:",inline"` // make them in root level not in BaseModel, by using an inline struct

	Email             string            `json:"email" bson:"email" validate:"required,email"`
	Phone             string            `json:"phone,omitempty" bson:"phone,omitempty" validate:"omitempty,e164"`
	Username          string            `json:"username" bson:"username" validate:"required,min=3,max=50"`
	FirstName         string            `json:"first_name" bson:"first_name" validate:"required,min=1,max=100"`
	LastName          string            `json:"last_name" bson:"last_name" validate:"required,min=1,max=100"`
	Password          string            `json:"-" bson:"password" validate:"required,min=8"`
	Location          *string           `json:"location,omitempty" bson:"location,omitempty"` // from IP address, use profile.location for user-defined location
	VerificationLevel VerificationLevel `json:"verification_level" bson:"verification_level"`
	Roles             []UserRole        `json:"roles" bson:"roles"`
	LastLoginAt       *time.Time        `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	BannedAt          *time.Time        `json:"ban_date,omitempty" bson:"ban_date,omitempty"`
	BanReason         string            `json:"ban_reason,omitempty" bson:"ban_reason,omitempty"`
	Preferences       UserPrefs         `json:"preferences" bson:"preferences"`
	Profile           UserProfile       `json:"profile" bson:"profile"`
	// 3. party authentication
	AuthProvider    AuthProvider `json:"auth_provider" bson:"auth_provider"`
	ProviderID      string       `json:"provider_id,omitempty" bson:"provider_id,omitempty"`
	OAuthSignedUpAt *time.Time   `json:"oauth_signed_up_at,omitempty" bson:"oauth_signed_up_at,omitempty"`
	TermsAccepted   bool         `json:"terms_accepted" bson:"terms_accepted"`
}

type UserPrefs struct {
	Theme              string `json:"theme" bson:"theme"`
	Language           string `json:"language" bson:"language"`
	Timezone           string `json:"timezone" bson:"timezone"`
	EmailNotifications bool   `json:"email_notifications" bson:"email_notifications"`
	PushNotifications  bool   `json:"push_notifications" bson:"push_notifications"`
}

type UserProfile struct {
	IsPublicProfile bool              `json:"is_public_profile" bson:"is_public_profile"`                     //default false                    // boolean
	InterestInPaccy string            `json:"interest_in_paccy,omitempty" bson:"interest_in_paccy,omitempty"` // Why are you interested in Paccy?
	ReferralSource  string            `json:"referral_source,omitempty" bson:"referral_source,omitempty"`     // How did you hear about us?
	AvatarURL       string            `json:"avatar_url" bson:"avatar_url"`
	Bio             string            `json:"bio" bson:"bio"`
	GraduationYear  int               `json:"graduation_year" bson:"graduation_year"`
	School          string            `json:"school" bson:"school"`
	Website         string            `json:"website" bson:"website"`
	SocialLinks     map[string]string `json:"social_links" bson:"social_links"`
	Birthdate       *time.Time        `json:"birthdate,omitempty" bson:"birthdate,omitempty"`
	Location        *string           `json:"location,omitempty" bson:"location,omitempty"` // User-defined location, can be different from IP-based location
}

type SafeUser struct {
	models.BaseModel                    // ID, UUID, CreatedAt, UpdatedAt, DeletedAt, Version
	Email             string            `json:"email"`
	Phone             string            `json:"phone,omitempty"`
	Username          string            `json:"username"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	Location          *string           `json:"location,omitempty"`
	VerificationLevel VerificationLevel `json:"verification_level"`
	Roles             []UserRole        `json:"roles"`
	LastLoginAt       *time.Time        `json:"last_login_at,omitempty"`
	BannedAt          *time.Time        `json:"banned_at,omitempty"`
	BanReason         string            `json:"ban_reason,omitempty"`
	Preferences       UserPrefs         `json:"preferences"`
	Profile           UserProfile       `json:"profile"`
}

func GetDefaultUserPrefs() UserPrefs {
	return UserPrefs{
		Theme:              "dark",
		Language:           "en",
		Timezone:           "UTC",
		EmailNotifications: true,
		PushNotifications:  true,
	}
}

func (u *User) BeforeCreate() {
	u.BaseModel.BeforeCreate()

	u.Username = strings.ToLower(strings.TrimSpace(u.Username))
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.DeletedAt = nil // Ensure DeletedAt is nil on creation

	u.Preferences = GetDefaultUserPrefs()

	if u.AuthProvider == "" {
		u.AuthProvider = ProviderEmail
	}

	if len(u.Roles) == 0 {
		u.Roles = []UserRole{RoleUser}
	}

}

func (u *User) BeforeUpdate() {
	u.BaseModel.BeforeUpdate()
}

func (u *User) GetFullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func GetCollectionName() string {
	return "users"
}

func (u *User) IsUserBanned() bool {
	return u.BannedAt != nil
}

// is account active
func (u *User) IsActive() bool {
	return !u.IsUserBanned() && u.DeletedAt == nil
}

func (u *User) IsEmailVerified() bool {
	return u.VerificationLevel >= EmailVerified && u.Email != ""
}

func (u *User) IsPhoneVerified() bool {
	return u.VerificationLevel >= PhoneVerified && u.Phone != ""
}

func (u *User) IsOAuthUser() bool {
	return u.AuthProvider != ProviderEmail
}

func (u *User) IsVerified() bool {
	return u.IsEmailVerified() || u.IsPhoneVerified()
}

func (u *User) IsTrustedUser() bool {
	return u.VerificationLevel >= TrustedUser
}

// ban user using setRoles
func (u *User) Ban(reason string) bool {
	now := time.Now().UTC()
	u.BannedAt = &now
	u.BanReason = reason

	// Set roles to empty to prevent further actions
	u.SetRoles([]UserRole{RoleBanned})
	u.VerificationLevel = Banned

	return u.IsUserBanned()
}

func (u *User) HasRole(role UserRole) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) AddRole(role UserRole) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

func (u *User) RemoveRole(role UserRole) {
	for i, r := range u.Roles {
		if r == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			break
		}
	}
}

func (u *User) SetRoles(roles []UserRole) { // removes all roles and sets new ones
	u.Roles = []UserRole{}
	for _, role := range roles {
		if IsValidRole(string(role)) {
			u.AddRole(role)
		}
	}
}

func (u *User) ToSafe() *SafeUser {
	return &SafeUser{
		BaseModel:         u.BaseModel,
		Email:             u.Email,
		Phone:             u.Phone,
		Username:          u.Username,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Location:          u.Location,
		VerificationLevel: u.VerificationLevel,
		Roles:             u.Roles,
		LastLoginAt:       u.LastLoginAt,
		BannedAt:          u.BannedAt,
		BanReason:         u.BanReason,
		Preferences:       u.Preferences,
		Profile:           u.Profile,
	} // emitted fields: Password, AuthProvider, ProviderID, OAuthSignedUpAt
}
