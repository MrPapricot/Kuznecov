package handlers

import (
	"github.com/gofiber/fiber/v3"

	shared "shared/auth"

	"github.com/google/uuid"
)

// Register обрабатывает POST /api/auth/register
func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest

	// Fiber сам парсит JSON в структуру
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "All fields are required",
		})
	}

	authReq := shared.AuthRequest{
		CorrelationID: uuid.New().String(),
		Type:          shared.RequestTypeRegister,
		Username:      req.Username,
		Email:         req.Email,
		Password:      req.Password,
	}

	// Используем UserContext() для получения context.Context из Fiber
	resp, err := h.kafkaClient.SendRequest(c.Context(), authReq, h.timeout)
	if err != nil {
		return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !resp.Success {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": resp.Error,
		})
	}

	// Возвращаем JSON с кодом 201 Created
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token":      resp.Token,
		"expires_at": resp.ExpiresAt,
		"user_id":    resp.UserID,
	})
}
