package dto

type CreateUserDTO struct {
	Email             string `json:"email" validate:"required,email"`
	Username          string `json:"username" validate:"required,alphanum,min=3,max=50"`
	FirstName         string `json:"first_name" validate:"required,min=1,max=100"`
	LastName          string `json:"last_name" validate:"required,min=1,max=100"`
	Password          string `json:"password" validate:"required,min=8"`
	AuthProvider      string `json:"auth_provider,omitempty"`
	ProviderID        string `json:"provider_id,omitempty"`
	TermsAccepted      bool    `json:"terms_accepted" validate:"required"`
}

type UpdateUserDTO struct { // all fields are optional; No password or auth provider updates or location
	Username          *string `json:"username,omitempty" validate:"omitempty,alphanum,min=3,max=50"`
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
}

type UpdateUserPreferencesDTO struct { // all fields are optional
	Theme              *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark"`
	Language           *string `json:"language,omitempty" validate:"omitempty,"` //oneof=en-US en-GB es-ES fr-FR de-DE it-IT"
	Timezone           *string `json:"timezone,omitempty" validate:"omitempty,timezone"`
	EmailNotifications *bool   `json:"email_notifications,omitempty" validate:"omitempty"`
	PushNotifications  *bool   `json:"push_notifications,omitempty" validate:"omitempty"`
}

type UpdateUserProfileDTO struct { // all fields are optional
	AvatarURL       *string           `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Bio             *string           `json:"bio,omitempty" validate:"omitempty,max=500"`
	GraduationYear  *int              `json:"graduation_year,omitempty" validate:"omitempty,gte=1900,lte=2100"`
	School          *string           `json:"school,omitempty" validate:"omitempty,max=100"`
	Website         *string           `json:"website,omitempty" validate:"omitempty,url"`
	SocialLinks     map[string]string `json:"social_links,omitempty" validate:"omitempty,dive,keys,required,max=50,alphanum,dive,required,url"`
	Birthdate       *string           `json:"birthdate,omitempty" validate:"omitempty,date"`
	Location        *string           `json:"location,omitempty" validate:"omitempty,max=100"` // User-defined location, can be different from IP-based location
	IsPublicProfile *bool   `json:"is_public_profile,omitempty" validate:"omitempty"` // boolean
}



	