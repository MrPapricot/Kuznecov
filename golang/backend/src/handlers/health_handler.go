package handlers

import "github.com/gofiber/fiber/v3"

func HealthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "OK"})
}
