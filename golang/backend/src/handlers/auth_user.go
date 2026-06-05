package handlers

import (
	shared "shared/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	authReq := shared.AuthRequest{
		CorrelationID: uuid.New().String(),
		Type:          shared.RequestTypeLogin,
		Email:         req.Email,
		Password:      req.Password,
	}

	resp, err := h.kafkaClient.SendRequest(c.Context(), authReq, h.timeout)
	if err != nil {
		return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": err.Error()})
	}
	if !resp.Success {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": resp.Error})
	}

	return c.JSON(fiber.Map{
		"token": resp.Token, "expires_at": resp.ExpiresAt, "user_id": resp.UserID,
	})
}
