package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/paccy-io/paccy-backend/pkg/logger"
)

const DefaultJWTSecretKey = "your_secret_key"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
	Hashing  HashingConfig
	Notify   NotifyConfig

	//  Redis 		RedisConfig

}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	URI      string
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int
	Issuer    string
}

type AppConfig struct {
	Name              string
	Version           string
	Environment       string
	LoggingLevel      string
	DiscordWebhookURL string
}

// argon2id parameters
type HashingConfig struct {
	Algorithm   string //"argon2id"
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type NotifyConfig struct {
	Project ProjectNotifyQuotaConfig
	User    UserNotifyQuotaConfig
	Admin   AdminNotifyConfig
}
type ResenderConfig struct {
	APIKey string
	From   string
}

type GmailSmtpConfig struct {
	Username string
	Password string
	From     string
	Host     string
	Port     string
}

type AdminNotifyConfig struct {
	Resender  ResenderConfig
	GmailSmtp GmailSmtpConfig
}

type ProjectNotifyQuotaConfig struct {
	ProjectName    string
	ResenderQuota  Quota
	GmailSmtpQuota Quota
}

type UserNotifyQuotaConfig struct {
	AuthQuota         int
	SecurityQuota     int
	NotificationQuota int
}

func LoadConfig() *Config {
	// Create config first to get log level and environment
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "3000"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			URI:      getEnv("DATABASE_URI", "mongodb://localhost:27017"),
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "27017"),
			Username: getEnv("DATABASE_USERNAME", "admin"),
			Password: getEnv("DATABASE_PASSWORD", "admin"),
			DBName:   getEnv("DATABASE_NAME", "paccy"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET_KEY", DefaultJWTSecretKey),
			ExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 3600*8), // default 8 hours in seconds
			Issuer:    getEnv("JWT_ISSUER", "paccy"),         // main issuer, but in-app auth providers can have their own issuers (ex: auth-service)
		},
		App: AppConfig{
			Name:              getEnv("APP_NAME", "Paccy"),
			Version:           getEnv("APP_VERSION", "1.0.0"),
			Environment:       getEnv("APP_ENVIRONMENT", "development"),
			LoggingLevel:      getEnv("APP_LOGGING_LEVEL", "info"),
			DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		},
		Hashing: HashingConfig{
			Algorithm:   getEnv("HASHING_ALGORITHM", "argon2id"),
			Memory:      uint32(getEnvAsInt("HASHING_MEMORY", 65536)),
			Iterations:  uint32(getEnvAsInt("HASHING_ITERATIONS", 3)),
			Parallelism: uint8(getEnvAsInt("HASHING_PARALLELISM", 1)),
			SaltLength:  uint32(getEnvAsInt("HASHING_SALT_LENGTH", 16)),
			KeyLength:   uint32(getEnvAsInt("HASHING_KEY_LENGTH", 32)),
		},
		Notify: NotifyConfig{
			Project: ProjectNotifyQuotaConfig{
				ProjectName: getEnv("NOTIFY_PROJECT_NAME", "paccy"),
				ResenderQuota: Quota{
					DailyQuota:   getEnvAsInt("NOTIFY_RESENDER_DAILY_QUOTA", 100),
					MonthlyQuota: getEnvAsInt("NOTIFY_RESENDER_MONTHLY_QUOTA", 3000),
				},
				GmailSmtpQuota: Quota{
					DailyQuota:   getEnvAsInt("NOTIFY_GMAIL_SMTP_DAILY_QUOTA", 450),
					MonthlyQuota: getEnvAsInt("NOTIFY_GMAIL_SMTP_MONTHLY_QUOTA", 450*31), // 450 emails per day, multiplied by 31 days for monthly quota
				},
			},
			User: UserNotifyQuotaConfig{
				AuthQuota:         getEnvAsInt("NOTIFY_USER_AUTH_QUOTA", 2),
				SecurityQuota:     getEnvAsInt("NOTIFY_USER_SECURITY_QUOTA", 1),
				NotificationQuota: getEnvAsInt("NOTIFY_USER_NOTIFICATION_QUOTA", 0),
			},
			Admin: AdminNotifyConfig{
				Resender: ResenderConfig{
					APIKey: getEnv("NOTIFY_ADMIN_RESENDER_API_KEY", ""),
					From:   getEnv("NOTIFY_ADMIN_RESENDER_FROM", "no-reply@paccy.io"),
				},
				GmailSmtp: GmailSmtpConfig{
					Username: getEnv("NOTIFY_ADMIN_GMAIL_SMTP_USERNAME", ""),
					Password: getEnv("NOTIFY_ADMIN_GMAIL_SMTP_PASSWORD", ""),
					From:     getEnv("NOTIFY_ADMIN_GMAIL_SMTP_FROM", "no-reply@paccy.io"),
					Host:     getEnv("NOTIFY_ADMIN_GMAIL_SMTP_HOST", "smtp.gmail.com"),
					Port:     getEnv("NOTIFY_ADMIN_GMAIL_SMTP_PORT", "465"),
				},
			},
		},
	}

	// Initialize logger with config values
	logger.Initialize(cfg.App.LoggingLevel, cfg.App.Environment, cfg.App.DiscordWebhookURL)

	configLogger := logger.Config()
	configLogger.Info("Configuration loaded successfully")

	return cfg
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func (cfg *Config) GetDatabaseURI() string {
	return cfg.Database.URI
}

func (cfg *Config) GetDatabaseName() string {
	if cfg == nil || cfg.Database.DBName == "" {
		return "paccy" // default database name if not set
	}
	return cfg.Database.DBName
}

func IsProduction(cfg *Config) bool {
	if cfg == nil || cfg.App.Environment == "" {
		return false // default to development if not set
	}
	return cfg.App.Environment == "production"
}

func (cfg *Config) GetResendConfig() (string, string) {
	if cfg == nil {
		return "", "" // return empty if not set
	}

	return cfg.Notify.Admin.Resender.APIKey, cfg.Notify.Admin.Resender.From
}

// host, port, username, password, from respectively
func (cfg *Config) GetGmailSmtpConfig() (string, string, string, string, string) {
	if cfg == nil {
		return "", "", "", "", "" // return empty if not set
	}
	return cfg.Notify.Admin.GmailSmtp.Host, cfg.Notify.Admin.GmailSmtp.Port,
		cfg.Notify.Admin.GmailSmtp.Username, cfg.Notify.Admin.GmailSmtp.Password,
		cfg.Notify.Admin.GmailSmtp.From
}

// auth, security, notification quotas respectively
func (cfg *Config) GetUserNotificationQuota() (int, int, int) {
	if cfg == nil {
		return 0, 0, 0 // return zero if not set
	}
	return cfg.Notify.User.AuthQuota, cfg.Notify.User.SecurityQuota, cfg.Notify.User.NotificationQuota
}

// daily, monthly quotas respectively for Resender
func (cfg *Config) GetResenderQuota() Quota {
	if cfg == nil {
		return Quota{}
	}
	return cfg.Notify.Project.ResenderQuota
}

// daily, monthly quotas respectively for Gmail SMTP
func (cfg *Config) GetGmailSmtpQuota() Quota {
	if cfg == nil {
		return Quota{}
	}
	return cfg.Notify.Project.GmailSmtpQuota
}

func init() {
	// Only load .env file in development (when not in Azure/production)
	// Azure App Service provides environment variables directly
	if os.Getenv("APP_ENVIRONMENT") != "production" && os.Getenv("AZURE_CLIENT_ID") == "" {
		err := godotenv.Load("cmd/.env")
		if err != nil {
			// Don't use logger here as it's not initialized yet
			// log.Println("No .env file found or error reading it (this is normal in production):", err)
		}
	}
}
