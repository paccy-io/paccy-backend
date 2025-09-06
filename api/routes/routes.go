package routes

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/paccy-io/paccy-backend/api/handlers"
	"github.com/paccy-io/paccy-backend/api/middleware"
	"github.com/paccy-io/paccy-backend/config"
	"github.com/paccy-io/paccy-backend/internal/academic"
	"github.com/paccy-io/paccy-backend/internal/auth"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/calendar"
	"github.com/paccy-io/paccy-backend/internal/college"
	"github.com/paccy-io/paccy-backend/internal/essay"
	"github.com/paccy-io/paccy-backend/internal/extracurriculars"
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
	collegeRepo := college.NewMongoCollegeRepository(db)
	essayRepo := essay.NewMongoEssayRepository(db)
	calendarRepo := calendar.NewMongoCalendarRepository(db)
	extracurricularRepo := extracurriculars.NewMongoExtracurricularRepository(db)
	academicRepo := academic.NewAcademicRepository(db)
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
	collegeService := college.NewCollegeService(collegeRepo, validate)
	essayService := essay.NewEssayService(essayRepo, collegeRepo, validate)
	calendarService := calendar.NewCalendarService(calendarRepo, collegeRepo, validate)
	extracurricularService := extracurriculars.NewExtracurricularService(extracurricularRepo)
	academicService := academic.NewAcademicService(academicRepo)
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
	collegeHandler := handlers.NewCollegeHandler(collegeService)
	essayHandler := handlers.NewEssayHandler(essayService)
	calendarHandler := handlers.NewCalendarHandler(calendarService)
	extracurricularHandler := handlers.NewExtracurricularHandler(extracurricularService)
	academicHandler := handlers.NewAcademicHandler(academicService)
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
	public.Post("/register", authHandler.RegisterAndSendVerification)
	public.Post("/login", authHandler.Login)

	// Notification routes
	//public.Post("/notify", notificationHandler.SendNotification)
	//public.Post("/notify/quota", notificationHandler.CreateProjectQuota)

	// User routes
	public.Get("/users", userHandler.GetUsers)

	// Protected user routes (require JWT + user ID validation + user existence check)
	public.Get("/users/:userID", middleware.CombinedJWTUserIDWithValidationMiddleware(authService, userService), userHandler.GetUser)
	public.Put("/users/:userID", middleware.CombinedJWTUserIDWithValidationMiddleware(authService, userService), userHandler.UpdateUserFields)
	public.Put("/users/:userID/profile", middleware.CombinedJWTUserIDWithValidationMiddleware(authService, userService), userHandler.UpdateProfile)
	public.Put("/users/:userID/preferences", middleware.CombinedJWTUserIDWithValidationMiddleware(authService, userService), userHandler.UpdatePreferences)

	// Protected college routes (require JWT authentication + user existence validation)
	public.Post("/college-applications", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.CreateCollegeApplication)
	public.Get("/college-applications", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetCollegeApplications)
	public.Get("/college-applications/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetCollegeApplication)
	public.Put("/college-applications/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.UpdateCollegeApplication)
	public.Delete("/college-applications/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.DeleteCollegeApplication)
	//public.Get("/college-applications/status/:status", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetCollegeApplicationsByStatus)
	//public.Get("/college-applications/priority/:priority", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetCollegeApplicationsByPriority)
	public.Get("/college-applications/deadlines/upcoming", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetUpcomingDeadlines)

	// Checklist routes
	public.Post("/college-applications/checklists", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.CreateChecklist)
	public.Get("/college-applications/:applicationId/checklists/:checklistId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetChecklist)
	public.Put("/college-applications/:applicationId/checklists/:checklistId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.UpdateChecklist)
	public.Delete("/college-applications/:applicationId/checklists/:checklistId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.DeleteChecklist)

	// Task routes
	public.Post("/college-applications/tasks", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.CreateTask)
	public.Get("/college-applications/:applicationId/checklists/:checklistId/tasks/:taskId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.GetTask)
	public.Put("/college-applications/:applicationId/checklists/:checklistId/tasks/:taskId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.UpdateTask)
	public.Delete("/college-applications/:applicationId/checklists/:checklistId/tasks/:taskId", middleware.JWTWithUserValidationMiddleware(authService, userService), collegeHandler.DeleteTask)

	// Essay routes
	public.Post("/essays", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.CreateEssay)
	public.Get("/essays", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetUserEssays)
	public.Get("/essays/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetEssay)
	public.Put("/essays/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.UpdateEssay)
	public.Delete("/essays/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.DeleteEssay)
	//public.Get("/essays/status", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetEssaysByStatus)
	//public.Get("/essays/deadlines", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetEssaysWithDeadlines) !! error
	//public.Get("/essays/count", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetEssayCount)
	public.Get("/college-applications/:college_application_id/essays", middleware.JWTWithUserValidationMiddleware(authService, userService), essayHandler.GetEssaysByCollegeApplication)

	// Calendar routes
	public.Post("/calendar", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.CreateEvent)
	public.Get("/calendar", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetUserEvents)
	public.Get("/calendar/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEvent)
	public.Put("/calendar/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.UpdateEvent)
	public.Delete("/calendar/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.DeleteEvent)
	//public.Get("/calendar/inbox", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetInboxEvents)
	//public.Get("/calendar/date-range", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEventsByDateRange)
	//public.Get("/calendar/type", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEventsByType)
	//public.Get("/calendar/priority", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEventsByPriority)
	//public.Get("/calendar/upcoming", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetUpcomingEvents)
	//public.Get("/calendar/count", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEventCount)
	public.Get("/college-applications/:college_application_id/calendar", middleware.JWTWithUserValidationMiddleware(authService, userService), calendarHandler.GetEventsByCollegeApplication)

	// Extracurricular routes
	public.Post("/extracurriculars", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.CreateExtracurricular)
	public.Get("/extracurriculars", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.GetExtracurriculars)
	public.Get("/extracurriculars/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.GetExtracurricularByID)
	public.Put("/extracurriculars/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.UpdateExtracurricular)
	public.Delete("/extracurriculars/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.DeleteExtracurricular)
	public.Get("/extracurriculars/category/:category", middleware.JWTWithUserValidationMiddleware(authService, userService), extracurricularHandler.GetExtracurricularsByCategory)

	// Academic routes - Exams
	public.Post("/exams", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.CreateExam)
	public.Get("/exams", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.GetExamsByUserID)
	public.Get("/exams/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.GetExamByID)
	public.Put("/exams/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.UpdateExam)
	public.Delete("/exams/:id", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.DeleteExam)
	//public.Get("/exams/type/:type", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.GetExamsByType)

	// Academic routes - High School GPA
	public.Post("/high-school-gpa", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.CreateHighSchoolGPA)
	public.Get("/high-school-gpa", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.GetHighSchoolGPAByUserID)
	public.Put("/high-school-gpa", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.UpdateHighSchoolGPA)
	public.Delete("/high-school-gpa", middleware.JWTWithUserValidationMiddleware(authService, userService), academicHandler.DeleteHighSchoolGPA)

}
