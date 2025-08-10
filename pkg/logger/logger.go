package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is a structured logger with service context
type Logger struct {
	*logrus.Entry
	service string
}

// Global logger instance
var globalLogger *logrus.Logger

// CustomFormatter provides clean service-prefixed logging
type CustomFormatter struct {
	TimestampFormat string
	ForceColors     bool
}

// Format renders a log entry with service prefix
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format(f.TimestampFormat)

	// Get service name from entry data
	service := "system"
	if svc, ok := entry.Data["service"]; ok {
		service = fmt.Sprintf("%v", svc)
	}

	// Color codes for different log levels
	var levelColor string
	if f.ForceColors {
		switch entry.Level {
		case logrus.DebugLevel:
			levelColor = "\033[36m" // Cyan
		case logrus.InfoLevel:
			levelColor = "\033[32m" // Green
		case logrus.WarnLevel:
			levelColor = "\033[33m" // Yellow
		case logrus.ErrorLevel:
			levelColor = "\033[31m" // Red
		case logrus.FatalLevel:
			levelColor = "\033[35m" // Magenta
		default:
			levelColor = "\033[37m" // White
		}
	}

	// Format: 2024-08-10 15:04:05 [auth] INFO: User login successful
	if f.ForceColors {
		fmt.Fprintf(b, "%s [\033[34m%s\033[0m] %s%s\033[0m: %s",
			timestamp,
			strings.ToUpper(service),
			levelColor,
			strings.ToUpper(entry.Level.String()),
			entry.Message,
		)
	} else {
		fmt.Fprintf(b, "%s [%s] %s: %s",
			timestamp,
			strings.ToUpper(service),
			strings.ToUpper(entry.Level.String()),
			entry.Message,
		)
	}

	// Add additional fields if present (excluding service)
	for key, value := range entry.Data {
		if key != "service" {
			fmt.Fprintf(b, " %s=%v", key, value)
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// DiscordWebhook represents a Discord webhook configuration
type DiscordWebhook struct {
	URL     string
	Enabled bool
	client  *http.Client
}

// DiscordPayload represents the JSON payload for Discord webhooks
type DiscordPayload struct {
	Content string `json:"content"`
	Embeds  []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Color       int    `json:"color"`
		Timestamp   string `json:"timestamp"`
		Fields      []struct {
			Name   string `json:"name"`
			Value  string `json:"value"`
			Inline bool   `json:"inline"`
		} `json:"fields"`
	} `json:"embeds"`
}

// Global Discord webhook instance
var discordWebhook *DiscordWebhook

// InitializeDiscordWebhook sets up Discord webhook logging for critical errors
func InitializeDiscordWebhook(webhookURL string) {
	if webhookURL != "" {
		discordWebhook = &DiscordWebhook{
			URL:     webhookURL,
			Enabled: true,
			client:  &http.Client{Timeout: 10 * time.Second},
		}
	}
}

// SendToDiscord sends critical logs to Discord webhook
func (d *DiscordWebhook) SendToDiscord(entry *logrus.Entry) {
	if !d.Enabled || d.URL == "" {
		return
	}

	// Only send ERROR and FATAL level logs to Discord
	if entry.Level != logrus.ErrorLevel && entry.Level != logrus.FatalLevel {
		return
	}

	service := "system"
	if svc, ok := entry.Data["service"]; ok {
		service = fmt.Sprintf("%v", svc)
	}

	color := 15158332 // Red color for errors
	if entry.Level == logrus.FatalLevel {
		color = 10038562 // Dark red for fatal
	}

	payload := DiscordPayload{
		Content: fmt.Sprintf("🚨 **%s Error Alert**", strings.ToUpper(service)),
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       fmt.Sprintf("[%s] %s", strings.ToUpper(service), strings.ToUpper(entry.Level.String())),
				Description: entry.Message,
				Color:       color,
				Timestamp:   entry.Time.Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{},
			},
		},
	}

	// Add additional fields
	for key, value := range entry.Data {
		if key != "service" && len(payload.Embeds[0].Fields) < 10 { // Limit fields
			payload.Embeds[0].Fields = append(payload.Embeds[0].Fields, struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			}{
				Name:   key,
				Value:  fmt.Sprintf("%v", value),
				Inline: true,
			})
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently to avoid log loops
	}

	go func() {
		resp, err := d.client.Post(d.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

// DiscordHook implements logrus.Hook interface
type DiscordHook struct {
	webhook *DiscordWebhook
}

// NewDiscordHook creates a new Discord hook
func NewDiscordHook(webhook *DiscordWebhook) *DiscordHook {
	return &DiscordHook{webhook: webhook}
}

// Fire sends the log entry to Discord
func (hook *DiscordHook) Fire(entry *logrus.Entry) error {
	if hook.webhook != nil {
		hook.webhook.SendToDiscord(entry)
	}
	return nil
}

// Levels returns the log levels this hook should fire for
func (hook *DiscordHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel}
}

// Initialize sets up the global logger with configuration
func Initialize(level, environment, discordWebhookURL string) {
	globalLogger = logrus.New()

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		globalLogger.SetLevel(logrus.DebugLevel)
	case "info":
		globalLogger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		globalLogger.SetLevel(logrus.WarnLevel)
	case "error":
		globalLogger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		globalLogger.SetLevel(logrus.FatalLevel)
	default:
		globalLogger.SetLevel(logrus.InfoLevel)
	}

	// Always use custom formatter for clean service-prefixed logs
	globalLogger.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     environment != "production", // Colors in development only
	})

	// Set output
	globalLogger.SetOutput(os.Stdout)

	// Initialize Discord webhook if provided
	if discordWebhookURL != "" {
		InitializeDiscordWebhook(discordWebhookURL)
		if discordWebhook != nil {
			globalLogger.AddHook(NewDiscordHook(discordWebhook))
		}
	}
}

// New creates a new logger for a specific service
func New(service string) *Logger {
	if globalLogger == nil {
		Initialize("info", "development", "") // fallback initialization
	}

	return &Logger{
		Entry:   globalLogger.WithField("service", service),
		service: service,
	}
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	return &Logger{
		Entry:   l.Entry.WithFields(fields),
		service: l.service,
	}
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Entry:   l.Entry.WithField(key, value),
		service: l.service,
	}
}

// WithUserID adds user ID context to logs
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithField("user_id", userID)
}

// WithRequestID adds request ID context to logs
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// WithError adds error context to logs
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Entry:   l.Entry.WithError(err),
		service: l.service,
	}
}

// Service-specific logger constructors
func Auth() *Logger {
	return New("auth")
}

func User() *Logger {
	return New("user")
}

func Notify() *Logger {
	return New("notify")
}

func Mongo() *Logger {
	return New("mongo")
}

func Server() *Logger {
	return New("server")
}

func Config() *Logger {
	return New("config")
}

// HTTP middleware logger
func HTTP() *Logger {
	return New("http")
}

// JWT logger
func JWT() *Logger {
	return New("jwt")
}

// SendStartupNotification sends a notification to Discord when the application starts
func SendStartupNotification(appName, version, environment string, port string) {
	if discordWebhook == nil || !discordWebhook.Enabled || discordWebhook.URL == "" {
		return
	}

	// Green color for successful startup
	color := 5763719 // Green

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	payload := DiscordPayload{
		Content: fmt.Sprintf("🚀 **%s Started Successfully**", appName),
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       fmt.Sprintf("🟢 %s Deployment", appName),
				Description: "Application has been deployed/restarted successfully",
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{
					{
						Name:   "Version",
						Value:  version,
						Inline: true,
					},
					{
						Name:   "Environment",
						Value:  environment,
						Inline: true,
					},
					{
						Name:   "Port",
						Value:  port,
						Inline: true,
					},
					{
						Name:   "Hostname",
						Value:  hostname,
						Inline: true,
					},
					{
						Name:   "Startup Time",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently
	}

	go func() {
		resp, err := discordWebhook.client.Post(discordWebhook.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

// SendUserRegistrationNotification sends a notification when a new user registers
func SendUserRegistrationNotification(userID, username, email string) {
	if discordWebhook == nil || !discordWebhook.Enabled || discordWebhook.URL == "" {
		return
	}

	// Blue color for user registration
	color := 3447003 // Blue

	payload := DiscordPayload{
		Content: fmt.Sprintf("👤 **New User Registration**"),
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       "🆕 User Registered",
				Description: "A new user has successfully registered",
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{
					{
						Name:   "Username",
						Value:  username,
						Inline: true,
					},
					{
						Name:   "Email", 
						Value:  email,
						Inline: true,
					},
					{
						Name:   "User ID",
						Value:  userID,
						Inline: true,
					},
					{
						Name:   "Registration Time",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: false,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently
	}

	go func() {
		resp, err := discordWebhook.client.Post(discordWebhook.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

// SendEmailVerificationNotification sends a notification when a user verifies their email
func SendEmailVerificationNotification(userID, email string) {
	if discordWebhook == nil || !discordWebhook.Enabled || discordWebhook.URL == "" {
		return
	}

	// Green color for successful verification
	color := 5763719 // Green

	payload := DiscordPayload{
		Content: fmt.Sprintf("✅ **Email Verified**"),
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       "📧 Email Verification Complete",
				Description: "User has successfully verified their email address",
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{
					{
						Name:   "Email",
						Value:  email,
						Inline: true,
					},
					{
						Name:   "User ID",
						Value:  userID,
						Inline: true,
					},
					{
						Name:   "Verification Time",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: false,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently
	}

	go func() {
		resp, err := discordWebhook.client.Post(discordWebhook.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

// SendEmailSuccessNotification sends a notification when an email is sent successfully
func SendEmailSuccessNotification(userID, email, subject, template, provider string) {
	if discordWebhook == nil || !discordWebhook.Enabled || discordWebhook.URL == "" {
		return
	}

	// Green color for successful email delivery
	color := 5763719 // Green

	payload := DiscordPayload{
		Content: "📧 **Email Sent Successfully**",
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       "✅ Email Delivery Success",
				Description: "Email has been delivered successfully",
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{
					{
						Name:   "Recipient",
						Value:  email,
						Inline: true,
					},
					{
						Name:   "Subject",
						Value:  subject,
						Inline: true,
					},
					{
						Name:   "Template",
						Value:  template,
						Inline: true,
					},
					{
						Name:   "Provider",
						Value:  strings.ToUpper(provider),
						Inline: true,
					},
					{
						Name:   "User ID",
						Value:  userID,
						Inline: true,
					},
					{
						Name:   "Sent At",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently
	}

	go func() {
		resp, err := discordWebhook.client.Post(discordWebhook.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}

// SendEmailFailureNotification sends a notification when an email fails to send
func SendEmailFailureNotification(userID, email, subject, template, provider, errorMsg string) {
	if discordWebhook == nil || !discordWebhook.Enabled || discordWebhook.URL == "" {
		return
	}

	// Orange/yellow color for email failures
	color := 16776960 // Yellow/Orange

	payload := DiscordPayload{
		Content: "⚠️ **Email Delivery Failed**",
		Embeds: []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Color       int    `json:"color"`
			Timestamp   string `json:"timestamp"`
			Fields      []struct {
				Name   string `json:"name"`
				Value  string `json:"value"`
				Inline bool   `json:"inline"`
			} `json:"fields"`
		}{
			{
				Title:       "❌ Email Delivery Failure",
				Description: "Failed to deliver email to recipient",
				Color:       color,
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []struct {
					Name   string `json:"name"`
					Value  string `json:"value"`
					Inline bool   `json:"inline"`
				}{
					{
						Name:   "Recipient",
						Value:  email,
						Inline: true,
					},
					{
						Name:   "Subject",
						Value:  subject,
						Inline: true,
					},
					{
						Name:   "Template",
						Value:  template,
						Inline: true,
					},
					{
						Name:   "Provider",
						Value:  strings.ToUpper(provider),
						Inline: true,
					},
					{
						Name:   "User ID",
						Value:  userID,
						Inline: true,
					},
					{
						Name:   "Error",
						Value:  errorMsg,
						Inline: false,
					},
					{
						Name:   "Failed At",
						Value:  time.Now().Format("2006-01-02 15:04:05 MST"),
						Inline: true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return // Fail silently
	}

	go func() {
		resp, err := discordWebhook.client.Post(discordWebhook.URL, "application/json", bytes.NewBuffer(jsonPayload))
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}()
}
