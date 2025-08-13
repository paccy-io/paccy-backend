package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/paccy-io/paccy-backend/api/handlers"
	"github.com/paccy-io/paccy-backend/api/middleware"
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
	// Add logging middleware
	api.Use(middleware.RequestIDMiddleware())
	api.Use(middleware.LoggerMiddleware())
	api.Use(cors.New(cors.Config{
		AllowOrigins: "https://paccy.io, https://www.paccy.io, http://localhost:3000",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

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
	/*notificationHandler := handlers.NewNotificationHandler(notifyService, dto.QuotaDTO{
		ResenderQuota: dto.Quota{
			DailyQuota:   cfg.GetResenderQuota().DailyQuota,
			MonthlyQuota: cfg.GetResenderQuota().MonthlyQuota,
		},
		GmailSmtpQuota: dto.Quota{
			DailyQuota:   cfg.GetGmailSmtpQuota().DailyQuota,
			MonthlyQuota: cfg.GetGmailSmtpQuota().MonthlyQuota,
		},
	})
	*/

	// check the host name, if it is from https://paccy.azurewebsites.net/* then redirect to https://api.paccy.io/*
	api.Use(func(c *fiber.Ctx) error {
		if c.Hostname() == "paccy.azurewebsites.net" {
			return c.Redirect("https://api.paccy.io"+c.OriginalURL(), fiber.StatusMovedPermanently)
		}
		return c.Next()
	})

	public := api.Group("/")
	// Auth routes
	public.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the Paccy API")
	})
	public.Get("/verify", authHandler.VerifyToken)
	public.Post("/verify-email/:token", authHandler.VerifyEmail)
	public.Post("/send-verification/:userID", authHandler.SendEmailVerification)
	public.Post("/", authHandler.RegisterAndSendVerification)
	public.Post("/login", authHandler.Login)

	// Notification routes
	//public.Post("/notify", notificationHandler.SendNotification)
	//public.Post("/notify/quota", notificationHandler.CreateProjectQuota)

	// User routes
	public.Get("/users", userHandler.GetUsers)

	// Protected user routes (require JWT + user ID validation)
	public.Get("/users/:userID", middleware.CombinedJWTUserIDMiddleware(authService), userHandler.GetUser)
	public.Put("/users/:userID", middleware.CombinedJWTUserIDMiddleware(authService), userHandler.UpdateUserFields)
	public.Put("/users/:userID/profile", middleware.CombinedJWTUserIDMiddleware(authService), userHandler.UpdateProfile)
	public.Put("/users/:userID/preferences", middleware.CombinedJWTUserIDMiddleware(authService), userHandler.UpdatePreferences)

}
