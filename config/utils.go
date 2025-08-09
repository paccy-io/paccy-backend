package config

type Quota struct {
	DailyQuota   int `json:"daily_quota"`
	MonthlyQuota int `json:"monthly_quota"`
}

