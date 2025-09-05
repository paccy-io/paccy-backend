package dto

type CreateUserDTO struct {
	Email         string `json:"email" validate:"required,email"`
	Username      string `json:"username" validate:"required,alphanum,min=3,max=50"`
	FirstName     string `json:"first_name" validate:"required,min=1,max=100"`
	LastName      string `json:"last_name" validate:"required,min=1,max=100"`
	Password      string `json:"password" validate:"required,min=8"`
	AuthProvider  string `json:"auth_provider,omitempty"`
	ProviderID    string `json:"provider_id,omitempty"`
	TermsAccepted bool   `json:"terms_accepted" validate:"required"`
}

type UpdateUserDTO struct { // all fields are optional; No password or auth provider updates or location
	Username  *string `json:"username,omitempty" validate:"omitempty,alphanum,min=3,max=50"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
}

type UpdateUserPreferencesDTO struct { // all fields are optional
	Theme              *string `json:"theme,omitempty" validate:"omitempty,oneof=light dark"`
	Language           *string `json:"language,omitempty" validate:"omitempty,oneof=en en-US en-GB es es-ES fr fr-FR de de-DE it it-IT tr tr-TR zh zh-CN zh-TW ja ja-JP ko ko-KR pt pt-BR pt-PT ru ru-RU ar ar-SA nl nl-NL sv sv-SE da da-DK no no-NO fi fi-FI pl pl-PL cs cs-CZ hu hu-HU el el-GR he he-IL hi hi-IN th th-TH vi vi-VN"`
	Timezone           *string `json:"timezone,omitempty" validate:"omitempty"`
	EmailNotifications *bool   `json:"email_notifications,omitempty"`
	PushNotifications  *bool   `json:"push_notifications,omitempty"`
}

type UpdateUserProfileDTO struct { // all fields are optional
	AvatarURL       *string           `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Bio             *string           `json:"bio,omitempty" validate:"omitempty,max=500"`
	GraduationYear  *int              `json:"graduation_year,omitempty" validate:"omitempty,gte=1900,lte=2100"`
	School          *string           `json:"school,omitempty" validate:"omitempty,max=100"`
	Website         *string           `json:"website,omitempty" validate:"omitempty, max=500"`
	SocialLinks     map[string]string `json:"social_links,omitempty"`
	Birthdate       *string           `json:"birthdate,omitempty" validate:"omitempty"`
	Location        *string           `json:"location,omitempty" validate:"omitempty,max=100"`
	IsPublicProfile *bool             `json:"is_public_profile,omitempty"`
}
