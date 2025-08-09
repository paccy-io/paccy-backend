package dto

type Quota struct {
	DailyQuota   int `json:"daily_quota"`
	MonthlyQuota int `json:"monthly_quota"`
}

type QuotaDTO struct {
	ResenderQuota  Quota `json:"resender"`
	GmailSmtpQuota Quota `json:"gmail_smtp"`
}

type UserQuotaDTO struct {
	AuthQuota         int `json:"auth_quota"`
	SecurityQuota     int `json:"security_quota"`
	NotificationQuota int `json:"notification_quota"`
}

type SendNotificationDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Template string `json:"template"`          // HTML template name
	Link     string `json:"verification_link"` // link to be used in the template ( verification, password reset, promotion etc. )
}
