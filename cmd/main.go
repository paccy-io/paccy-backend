package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/api/routes"
	"github.com/paccy-io/paccy-backend/config"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/database"
	"github.com/paccy-io/paccy-backend/pkg/logger"
	"github.com/paccy-io/paccy-backend/pkg/stats"
)

func main() {
	cfg := config.LoadConfig()

	expIn := time.Duration(cfg.JWT.ExpiresIn) * time.Second
	jwtProvider := jwt.NewJWTProvider(cfg.JWT.SecretKey, expIn, cfg.JWT.Issuer)
	mongoDB, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	go stats.StartMemoryLogger(5 * time.Minute)

	app := fiber.New(fiber.Config{
		AppName:      "Paccy Backend API",
		ServerHeader: "Paccy",
		//ErrorHandler: middleware.ErrorHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world!")
	})

	api := app.Group("/api")
	routes.SetupRoutes(api, mongoDB.GetDatabase(), jwtProvider, cfg)

	// Send startup notification to Discord
	logger.SendStartupNotification(cfg.App.Name, cfg.App.Version, cfg.App.Environment, cfg.Server.Port)

	serverLogger := logger.Server()
	serverLogger.WithField("port", cfg.Server.Port).WithField("environment", cfg.App.Environment).Info("Starting server")

	log.Fatal(app.Listen(":" + cfg.Server.Port))

}
