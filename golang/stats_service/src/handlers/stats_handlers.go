package handler

import (
	"errors"
	"fmt"
	"stats_service/src/repository"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	repository repository.StatsRepository
}

func NewHandler(repository repository.StatsRepository) Handler {
	return Handler{repository: repository}
}

// GetGlobalStatsHandler GET /global
func (h *Handler) GetGlobalStatsHandler(ctx fiber.Ctx) error {
	stats, err := h.repository.GetGlobalStats()
	if err != nil {
		fmt.Printf("Database error: %#v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	}

	return ctx.JSON(fiber.Map{
		"total_users":      stats.TotalUsers,
		"total_rooms":      stats.TotalRooms,
		"total_characters": stats.TotalCharacters,
	})
}

// GetUserStatsHandler GET /user
func (h *Handler) GetUserStatsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	stats, err := h.repository.GetUserStats(token)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{
		"created_at":         stats.CreatedAt.UTC().Format(time.RFC3339),
		"owned_chars_count":  stats.OwnedCharsCount,
		"owned_rooms_count":  stats.OwnedRoomsCount,
		"joined_rooms_count": stats.JoinedRoomsCount,
		"top_character":      stats.TopCharacter, // Автоматически станет null, если nil
		"top_room":           stats.TopRoom,      // Автоматически станет null, если nil
	})
}

func (h *Handler) handleError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.InvalidToken):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
	case errors.Is(err, repository.TokenExpired):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
	case errors.Is(err, repository.NoUserFound):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
	case errors.Is(err, repository.DatabaseError):
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	}
}
