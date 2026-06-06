package handlers

import (
	"errors"
	"fmt"
	"user_service/src/repository"

	"github.com/gofiber/fiber/v3"
)

func (handler *Handler) ChangePasswordHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided through headers"})
	}

	// Парсим тело запроса
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "old_password and new_password are required"})
	}

	err := handler.repository.ChangePassword(token, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, repository.InvalidToken):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
		case errors.Is(err, repository.TokenExpired):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
		case errors.Is(err, repository.NoUserFound):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
		case errors.Is(err, repository.PasswordMissmath):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Incorrect old password"})
		case errors.Is(err, repository.DatabaseError):
			fmt.Printf("Database error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		default:
			fmt.Printf("Base error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		}
	}

	return ctx.JSON(fiber.Map{"message": "Password changed successfully"})
}
