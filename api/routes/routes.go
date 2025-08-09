package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/api/handlers"
	"github.com/paccy-io/paccy-backend/config"
	"github.com/paccy-io/paccy-backend/internal/auth"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/notify"
	"github.com/paccy-io/paccy-backend/internal/notify/dto"
	"github.com/paccy-io/paccy-backend/internal/notify/providers"
	"github.com/paccy-io/paccy-backend/internal/notify/repositories"
	"github.com/paccy-io/paccy-backend/internal/user"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(api fiber.Router, db *mongo.Database, jwtProvider jwt.Provider, cfg *config.Config) {
	// Initialize repositories
	s1, s2, s3 := cfg.GetUserNotificationQuota()

	userRepo := user.NewMongoUserRepository(db)
	userQuotaRepo := repositories.NewMongoUserQuotaRepository(db, &dto.UserQuotaDTO{
		AuthQuota:         s1,
		SecurityQuota:     s2,
		NotificationQuota: s3,
	})
	projectQuotaRepo := repositories.NewMongoProjectQuotaRepository(db, cfg.Notify.Project.ProjectName)

	// Initialize validation
	validate := validator.New()

	// Initialize providers

	resendProvider := providers.NewResendProvider(
		cfg.GetResendConfig(),
	)
	gmailProvider := providers.NewGmailSmtpProvider(
		cfg.GetGmailSmtpConfig(),
	)

	// Initialize services
	userService := user.NewUserService(userRepo, validate)
	notifyService := notify.NewService(
		resendProvider,
		gmailProvider,
		validate,
		userQuotaRepo,
		projectQuotaRepo,
	)
	authService := auth.NewAuthService(validate, jwtProvider, userService, notifyService)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)
	notificationHandler := handlers.NewNotificationHandler(notifyService, dto.QuotaDTO{
		ResenderQuota: dto.Quota{
			DailyQuota:   cfg.GetResenderQuota().DailyQuota,
			MonthlyQuota: cfg.GetResenderQuota().MonthlyQuota,
		},
		GmailSmtpQuota: dto.Quota{
			DailyQuota:   cfg.GetGmailSmtpQuota().DailyQuota,
			MonthlyQuota: cfg.GetGmailSmtpQuota().MonthlyQuota,
		},
	})

	public := api.Group("/")
	public.Get("/", userHandler.GetUsers)
	// Auth routes
	public.Get("/verify", authHandler.VerifyToken)
	public.Post("/verify-email/:token", authHandler.VerifyEmail)
	public.Post("/send-verification/:userID", authHandler.SendEmailVerification)
	public.Post("/", authHandler.RegisterAndSendVerification)
	public.Post("/login", authHandler.Login)

	// Notification routes
	public.Post("/notify", notificationHandler.SendNotification)
	public.Post("/notify/quota", notificationHandler.CreateProjectQuota)

	// User routes
	//	public.Get("/:id", userHandler.GetUser)

}
