package handlers

import (
	"errors"
	"fmt"
	"user_service/src/repository"

	"github.com/gofiber/fiber/v3"
)

func (handler *Handler) ChangeUsernameHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided through headers"})
	}

	// Парсим тело запроса
	var req struct {
		NewUsername string `json:"new_username"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.NewUsername == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "new_username is required"})
	}

	old_username, new_username, err := handler.repository.ChangeUsername(token, req.NewUsername)
	if err != nil {
		switch {
		case errors.Is(err, repository.InvalidToken):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
		case errors.Is(err, repository.TokenExpired):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
		case errors.Is(err, repository.NoUserFound):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
		case errors.Is(err, repository.UsernameAlreadyExists):
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{"Error": "This username is already used"})
		case errors.Is(err, repository.DatabaseError):
			fmt.Printf("Database error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		default:
			fmt.Printf("Base error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		}
	}

	return ctx.JSON(fiber.Map{
		"old_username": old_username,
		"new_username": new_username,
	})
}
