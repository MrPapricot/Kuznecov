package handlers

import "github.com/gofiber/fiber/v3"

func HelloHandler(c fiber.Ctx) error {
	return c.SendString("Hello user YA tvoy rot ebal")
}
