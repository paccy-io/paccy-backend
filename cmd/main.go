package main

import (
	"fmt"
	"log"
	"os"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/api/routes"
	"github.com/paccy-io/paccy-backend/config"
	"github.com/paccy-io/paccy-backend/internal/auth/jwt"
	"github.com/paccy-io/paccy-backend/internal/database"
	"github.com/paccy-io/paccy-backend/pkg/stats"
)

func main() {
	cfg := config.LoadConfig()

	// check env variables and print them
	fmt.Println("env", os.Getenv("NOTIFY_ADMIN_GMAIL_SMTP_USERNAME"))
	fmt.Println("mongo", os.Getenv("DATABASE_URI"), cfg.GetDatabaseURI())
	fmt.Println("last version")

	//print all env variables
	// Initilialize JWT provider
	// seconds to time.Duration
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
		//	ErrorHandler: middleware.ErrorHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world!")
	})

	api := app.Group("/api")
	routes.SetupRoutes(api, mongoDB.GetDatabase(), jwtProvider, cfg)

	log.Fatal(app.Listen(":3000"))

}
