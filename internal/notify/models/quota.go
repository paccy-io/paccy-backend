package models

type QuotaType string

const (
	QuotaTypeAuth         QuotaType = "auth"
	QuotaTypeSecurity     QuotaType = "security"
	QuotaTypeNotification QuotaType = "notification"
)
