package handlers

import (
	"errors"
	"time"
	"user_service/src/repository"

	"fmt"

	"github.com/gofiber/fiber/v3"
)

func (handler *Handler) GetUserInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided through headers"})
	}
	user_info, err := handler.repository.GetUserInfo(token)
	if err != nil {
		switch {
		case errors.Is(err, repository.InvalidToken):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
		case errors.Is(err, repository.TokenExpired):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
		case errors.Is(err, repository.NoUserFound):
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
		case errors.Is(err, repository.DatabaseError):
			fmt.Printf("Database error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		default:
			fmt.Printf("Base error: %#v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
		}
	}
	return ctx.JSON(fiber.Map{"email": user_info.Email, "username": user_info.UserName, "created_at": user_info.CreatedAt.UTC().Format(time.RFC3339)})
}
