package handlers

import (
	"errors"
	"fmt"
	"user_service/src/repository"

	"github.com/gofiber/fiber/v3"
)

func (handler *Handler) DeleteUserHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided through headers"})
	}

	// Парсим тело запроса
	var req struct {
		Password string `json:"password"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.Password == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "password is required"})
	}

	err := handler.repository.DeleteUser(token, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, repository.InvalidToken):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
		case errors.Is(err, repository.TokenExpired):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
		case errors.Is(err, repository.NoUserFound):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
		case errors.Is(err, repository.PasswordMissmath):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Incorrect password"})
		case errors.Is(err, repository.DatabaseError):
			fmt.Printf("Database error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		default:
			fmt.Printf("Base error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		}
	}

	return ctx.JSON(fiber.Map{"message": "User was deleted successfully"})
}
