package handlers

import (
	"MainBackend/src/user_client"

	"github.com/gofiber/fiber/v3"
)

type UserHandlers struct {
	user_service_client *user_client.UserServiceClient
}

func NewUserHandler(user_service_client *user_client.UserServiceClient) *UserHandlers {
	return &UserHandlers{
		user_service_client: user_service_client,
	}
}

// GetUserInfoHandler godoc
// @Summary Получить информацию о пользователе
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handlers.UserInfoResponse "Данные пользователя"
// @Failure 401 {object} handlers.ErrorResponse "Неавторизован"
// @Router /api/user/info [get]
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

// ChangeUsernameHandler godoc
// @Summary Изменить имя пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.ChangeUsernameRequest true "Новое имя"
// @Success 200 {object} handlers.ChangeUsernameResponse "Успешное изменение"
// @Failure 409 {object} handlers.ErrorResponse "Имя уже занято"
// @Router /api/user/change-username [post]
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

// ChangePasswordHandler godoc
// @Summary Изменить пароль пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.ChangePasswordRequest true "Пароли"
// @Success 200 {object} handlers.MessageResponse "Успешное изменение"
// @Failure 400 {object} handlers.ErrorResponse "Неверный старый пароль"
// @Router /api/user/change-password [post]
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

// DeleteUserHandler godoc
// @Summary Удалить аккаунт пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.DeleteUserRequest true "Пароль для подтверждения"
// @Success 200 {object} handlers.MessageResponse "Успешное удаление"
// @Failure 400 {object} handlers.ErrorResponse "Неверный пароль"
// @Router /api/user/delete [post]
func (handler *UserHandlers) DeleteUserHandler(ctx fiber.Ctx) error {
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

	resp, err := handler.user_service_client.DeleteUser(token, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "User service unavailable"})
	}

	return ctx.Status(resp.StatusCode).Send(resp.Body)
}
