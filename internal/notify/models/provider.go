package models

type ProviderType string

const (
	ProviderTypeResender  ProviderType = "resender"
	ProviderTypeGmailSmtp ProviderType = "gmail_smtp"
)
