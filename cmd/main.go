package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/paccy-io/paccy-backend/config"
	"github.com/paccy-io/paccy-backend/internal/database"
)

func main() {
	cfg := config.LoadConfig()

	mongoDB, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	app := fiber.New()


	app.Get("/", func(c *fiber.Ctx) error {
		fmt.Println(mongoDB.GetDatabase().Name())



		return c.SendString("Hello world!")
	})

	log.Fatal(app.Listen(":3000"))

}