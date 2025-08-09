package models

import (
	"errors"
	"time"

	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/models"
)

// this document will be created only once for the project, there will be 2 providers (resender and gmail smtp) and each provider will have its own quota

type Base struct {
	Quota     int       `bson:"quota" json:"quota"`
	Usage     int       `bson:"usage" json:"usage"`
	ResetTime time.Time `bson:"reset_time" json:"reset_time"`
}

type BaseProviderQuota struct {
	Daily        Base         `bson:"daily" json:"daily"`
	Monthly      Base         `bson:"monthly" json:"monthly"`
	IsBlocked    bool         `bson:"is_blocked" json:"is_blocked"`
	ProviderType ProviderType `bson:"provider_type" json:"provider_type"`
}

type ProjectQuota struct {
	models.BaseModel `bson:",inline"`  // embedded base model for common fields
	ProjectName      string            `bson:"project_name" json:"project_name"` // project identifier
	Resender         BaseProviderQuota `bson:"resender" json:"resender"`
	GmailSmtp        BaseProviderQuota `bson:"gmail_smtp" json:"gmail_smtp"`
}

func (q *ProjectQuota) BeforeCreate(resenderDailyQ, resenderMonthlyQ int, gmailSmtpDailyQ int, gmailSmtpMonthlyQ int) {
	q.BaseModel.BeforeCreate()                                                          // ensure base model fields are set
	q.Resender.BeforeCreate(resenderDailyQ, resenderMonthlyQ, ProviderTypeResender)     // initialize Resender quota with default values
	q.GmailSmtp.BeforeCreate(gmailSmtpDailyQ, gmailSmtpMonthlyQ, ProviderTypeGmailSmtp) // initialize GmailSmtp quota with default values
}

func (b *BaseProviderQuota) BeforeCreate(dailyQ, monthlyQ int, providerType ProviderType) {
	b.Daily.BeforeCreate(dailyQ, 24*time.Hour)        // initialize daily quota with default values
	b.Monthly.BeforeCreate(monthlyQ, 30*24*time.Hour) // initialize monthly quota with default values
	b.IsBlocked = false                               // default to not blocked
	b.ProviderType = providerType                     // set the provider type
}

func (b *Base) BeforeCreate(defaultQuota int, defaultLimit time.Duration) {
	if b.ResetTime.IsZero() {
		b.ResetTime = time.Now().UTC().Add(defaultLimit) // set default reset time to 24 hours from now
	}

	if b.Quota == 0 {
		b.Quota = defaultQuota // default quota, can be adjusted as needed
	}

	b.Usage = 0 // initialize usage to zero

}

func (b *Base) Reset(limit time.Duration) {
	b.ResetTime = time.Now().UTC().Add(limit) // reset the reset time to the specified limit
	b.Usage = 0                               // reset usage to zero
}

// Is Reset Needed func
func (b *Base) IsResetNeeded() bool {
	return time.Now().UTC().After(b.ResetTime) // check if the current time is after the reset time
}

func (q *BaseProviderQuota) DailyReset(limit time.Duration) {
	q.Daily.Reset(limit) // reset the daily quota
}

func (q *BaseProviderQuota) MonthlyReset(limit time.Duration) {
	q.Monthly.Reset(limit) // reset the monthly quota
}

func GetCollectionName() string {
	return "project_quota"
}

func (q *ProjectQuota) GetActiveProvider() ProviderType {
	if !q.Resender.IsBlocked && !q.Resender.IsQuotaExceeded() {
		return q.Resender.ProviderType
	}

	if !q.GmailSmtp.IsBlocked && !q.GmailSmtp.IsQuotaExceeded() {
		return q.GmailSmtp.ProviderType
	}

	return ""
}

func (b *BaseProviderQuota) RemoveBlocked() {
	b.IsBlocked = false // reset the blocked status
}

func (b *Base) IsQuotaExceeded() (bool, error) {
	if b.IsResetNeeded() {
		return false, ErrResetNeeded
	}

	return b.Usage >= b.Quota, nil
}

func (q *BaseProviderQuota) IsQuotaExceeded() bool {
	dailyExceeded, err := q.Daily.IsQuotaExceeded()

	if errors.Is(err, ErrResetNeeded) {
		q.DailyReset(24 * time.Hour) // reset daily quota if needed
	}

	monthlyExceeded, err := q.Monthly.IsQuotaExceeded()
	if errors.Is(err, ErrResetNeeded) {
		q.MonthlyReset(30 * 24 * time.Hour) // reset monthly quota if needed
	}

	return dailyExceeded || monthlyExceeded
}

// switch provider when quota is exceeded
func (q *ProjectQuota) SwitchProvider() ProviderType {
	if !q.Resender.IsBlocked && !q.Resender.IsQuotaExceeded() {
		return q.Resender.ProviderType
	}

	if !q.GmailSmtp.IsBlocked && !q.GmailSmtp.IsQuotaExceeded() {
		return q.GmailSmtp.ProviderType
	}

	return ""
}

func (b *BaseProviderQuota) IncreaseUsage() {
	if !b.IsBlocked && !b.IsQuotaExceeded() {
		b.Daily.Usage++
		b.Monthly.Usage++
	}

}
