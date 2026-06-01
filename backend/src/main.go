package main

import (
	"MainBackend/src/handlers"

	db_adapter "db_adapter/src"

	"github.com/gofiber/fiber/v3"
)

type User struct {
	name  string
	email string
}

type Response struct {
	Message string `json:"message"`
}

func main() {
	_ = db_adapter.PostgresAdapter{}

	app := fiber.New()

	app.Get("/", handlers.HelloHandler)

	app.Get("/health", handlers.HealthHandler)

	app.Get("/hello", func(c fiber.Ctx) error {
		response := Response{
			Message: "Hello",
		}
		return c.JSON(response)
	})

	app.Listen(":8000")
}
