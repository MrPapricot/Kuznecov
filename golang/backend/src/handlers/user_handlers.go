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
// @Description Возвращает username, email и дату создания аккаунта
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object "Данные пользователя" example({"username": "johndoe", "email": "john@example.com", "created_at": "2024-06-05T12:00:00Z"})
// @Failure 401 {object} object "Отсутствует или невалидный токен"
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
// @Description Обновляет username. Возвращает старое и новое имя.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object true "Новое имя пользователя" example({"new_username": "new_cool_name"})
// @Success 200 {object} object "Успешное изменение" example({"old_username": "johndoe", "new_username": "new_cool_name"})
// @Failure 400 {object} object "Новое имя не указано" example({"Error": "new_username is required"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 409 {object} object "Имя уже занято" example({"Error": "Username already exists"})
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

// DeleteUserHandler godoc
// @Summary Удалить аккаунт пользователя
// @Description Полностью удаляет пользователя и все его связи. Требует подтверждения текущим паролем.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object true "Пароль для подтверждения удаления" example({"password": "securePassword123"})
// @Success 200 {object} object "Успешное удаление" example({"message": "User deleted successfully"})
// @Failure 400 {object} object "Неверный пароль" example({"Error": "Incorrect old password"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Router /api/user/delete [post]
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
// @Description Полностью удаляет пользователя и все его связи. Требует подтверждения текущим паролем.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.DeleteUserRequest true "Пароль для подтверждения удаления"
// @Success 200 {object} object "Успешное удаление" example({"message": "User deleted successfully"})
// @Failure 400 {object} object "Неверный пароль" example({"Error": "Incorrect old password"})
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
