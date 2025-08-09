package models

// quota model

import (
	"time"

	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	. "github.com/paccy-io/paccy-backend/pkg/errors"
	"github.com/paccy-io/paccy-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ResetDuration = 24 * time.Hour // default reset duration for quotas

// Quota represents the quota model for notifications

type BaseQuota struct { // no daily or monthly, just a single quota (most of the time, daily quota, sometimes specific duration)
	Quota     int       `bson:"quota" json:"quota"`
	Usage     int       `bson:"usage" json:"usage"`
	ResetTime time.Time `bson:"reset_time" json:"reset_time"`
	IsBlocked bool      `bson:"is_blocked" json:"is_blocked"` // this is a permanent field, unlike the IsBlocked field in ProjectQuota which is temporary and can be reset after the reset time
}

type UserQuota struct {
	models.BaseModel `bson:",inline"`   // embedded base model for common fields
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id"`
	Auth             BaseQuota          `bson:"auth" json:"auth"`
	Security         BaseQuota          `bson:"security" json:"security"`
	Notification     BaseQuota          `bson:"notification" json:"notification"`
}

func (q *UserQuota) BeforeCreate(quota *dto.UserQuotaDTO) { // quotas will be initialized in service layer using cfg
	q.BaseModel.BeforeCreate() // ensure base model fields are set
	q.Auth.BeforeCreate(ResetDuration, quota.AuthQuota)
	q.Security.BeforeCreate(ResetDuration, quota.SecurityQuota)
	q.Notification.BeforeCreate(ResetDuration, quota.NotificationQuota)
}

func (b *BaseQuota) BeforeCreate(limit time.Duration, quota int) {
	b.ResetTime = time.Now().UTC().Add(limit)
	b.Usage = 0
	b.Quota = quota
	b.IsBlocked = false
}

func (q *UserQuota) BeforeUpdate() {
	q.BaseModel.BeforeUpdate()
}

func (b *BaseQuota) Reset(limit time.Duration) {
	b.ResetTime = time.Now().UTC().Add(limit)
	b.Usage = 0
}

func (b *BaseQuota) IsResetNeeded() bool {
	// check if the current time is past the reset time
	return time.Now().UTC().After(b.ResetTime)
}

func (b *BaseQuota) CanUse() (bool, error) {
	// check is reset needed
	if b.IsBlocked {
		return false, ErrUserBanned
	}

	if b.IsResetNeeded() {
		return true, ErrResetNeeded
	}

	// check if the usage exceeds the quota
	return b.Usage < b.Quota, nil
}

func (b *BaseQuota) Block() {
	b.IsBlocked = true
}

func (b *BaseQuota) Unblock() {
	b.IsBlocked = false
}

func (b *BaseQuota) IsQuotaBlocked() bool {
	return b.IsBlocked
}

// returns true if the usage was incremented successfully, false if it was not allowed
func (b *BaseQuota) IncrementUsage() bool {
	canUse, err := b.CanUse()

	if err == ErrResetNeeded {
		b.Reset(ResetDuration) // reset the quota if needed
	} else if err != nil || !canUse {
		return false // unexpected error, do not increment usage
	}

	b.Usage++

	return true
}

func (b *BaseQuota) DecrementUsage() bool {
	if b.Usage <= 0 {
		return false // cannot decrement usage below 0
	}

	b.Usage--

	if b.Usage == 0 {
		b.Reset(ResetDuration) // reset the quota if usage is 0
	}

	return true
}

func GetUserQuotaCollectionName() string {
	return "user_quotas"
}
