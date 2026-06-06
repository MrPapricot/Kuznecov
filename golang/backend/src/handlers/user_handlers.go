package handlers

import (
	"MainBackend/src/user_client"

	"github.com/gofiber/fiber/v3"
)

type UserHandlers struct {
	user_service_client *user_client.UserServiceClient
}

func NewUserHandlers(user_service_client *user_client.UserServiceClient) *UserHandlers {
	return &UserHandlers{
		user_service_client: user_service_client,
	}
}

// GetUserInfoHandler GET /api/user/info
func (handler *UserHandlers) GetUserInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided through headers"})
	}

	resp, err := handler.user_service_client.GetUserInfo(token)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "User service unavailable"})
	}

	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// ChangeUsernameHandler POST /api/user/change-username
func (handler *UserHandlers) ChangeUsernameHandler(ctx fiber.Ctx) error {
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

	resp, err := handler.user_service_client.ChangeUsername(token, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "User service unavailable"})
	}

	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// ChangePasswordHandler POST /api/user/change-password
func (handler *UserHandlers) ChangePasswordHandler(ctx fiber.Ctx) error {
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

	resp, err := handler.user_service_client.ChangePassword(token, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "User service unavailable"})
	}

	return ctx.Status(resp.StatusCode).Send(resp.Body)
}
