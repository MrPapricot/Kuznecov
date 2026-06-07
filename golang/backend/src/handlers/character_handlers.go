package handlers

import (
	client "MainBackend/src/character_client"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CharacterHandlers struct {
	charServiceClient *client.CharacterServiceClient
}

func NewCharacterHandlers(charServiceClient *client.CharacterServiceClient) *CharacterHandlers {
	return &CharacterHandlers{
		charServiceClient: charServiceClient,
	}
}

// CreateCharacterHandler godoc
// @Summary Создать нового персонажа
// @Description Создает персонажа. Сумма характеристик должна быть равна 40 + (level - 1). Каждая характеристика от 0 до 10.
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object true "Данные персонажа" example({"name": "Vault Dweller", "description": "A brave survivor", "level": 1, "strength": 5, "perception": 5, "endurance": 5, "charisma": 5, "intelligence": 5, "agility": 5, "luck": 10})
// @Success 201 {object} object "UUID созданного персонажа" example({"char_uuid": "550e8400-e29b-41d4-a716-446655440000"})
// @Failure 400 {object} object "Ошибка валидации данных или суммы очков" example({"Error": "Invalid stat points sum. Must equal 40 + (level - 1)"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 502 {object} object "Сервис персонажей недоступен" example({"Error": "Character service unavailable"})
// @Router /api/character/create [post]
func (h *CharacterHandlers) CreateCharacterHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	fmt.Printf("🔍 RAW BODY: %s\n", string(ctx.Body()))

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Level        *int16 `json:"level"`
		Strength     *int16 `json:"strength"`
		Perception   *int16 `json:"perception"`
		Endurance    *int16 `json:"endurance"`
		Charisma     *int16 `json:"charisma"`
		Intelligence *int16 `json:"intelligence"`
		Agility      *int16 `json:"agility"`
		Luck         *int16 `json:"luck"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		fmt.Printf("Error Binding in create user: %#v\n", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.Name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "name is required"})
	}

	resp, err := h.charServiceClient.CreateCharacter(token, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// UpdateNameHandler godoc
// @Summary Изменить имя персонажа
// @Description Обновляет имя персонажа. Доступно только владельцу.
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body object true "Новое имя" example({"new_name": "New Vault Dweller"})
// @Success 200 {object} object "Старое и новое имя" example({"old_name": "Vault Dweller", "new_name": "New Vault Dweller"})
// @Failure 400 {object} object "Неверный формат UUID или тела запроса" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является владельцем" example({"Error": "User is not the owner of this character"})
// @Failure 404 {object} object "Персонаж не найден" example({"Error": "Character not found"})
// @Router /api/character/update-name/{uuid} [patch]
func (h *CharacterHandlers) UpdateNameHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(charUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewName string `json:"new_name"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.NewName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "new_name is required"})
	}

	resp, err := h.charServiceClient.UpdateName(token, charUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// UpdateDescriptionHandler godoc
// @Summary Изменить описание персонажа
// @Description Обновляет описание персонажа. Доступно только владельцу.
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body object true "Новое описание" example({"new_description": "A seasoned wasteland traveler"})
// @Success 200 {object} object "Старое и новое описание" example({"old_description": "...", "new_description": "A seasoned wasteland traveler"})
// @Failure 400 {object} object "Неверный формат UUID или тела запроса" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является владельцем" example({"Error": "User is not the owner of this character"})
// @Failure 404 {object} object "Персонаж не найден" example({"Error": "Character not found"})
// @Router /api/character/update-description/{uuid} [patch]
func (h *CharacterHandlers) UpdateDescriptionHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(charUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewDescription string `json:"new_description"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	resp, err := h.charServiceClient.UpdateDescription(token, charUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// DeleteCharacterHandler godoc
// @Summary Удалить персонажа
// @Description Полностью удаляет персонажа. Доступно только владельцу.
// @Tags Character
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Success 200 {object} object "Сообщение об успехе" example({"message": "Character deleted successfully"})
// @Failure 400 {object} object "Неверный формат UUID" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является владельцем" example({"Error": "User is not the owner of this character"})
// @Failure 404 {object} object "Персонаж не найден" example({"Error": "Character not found"})
// @Router /api/character/{uuid} [delete]
func (h *CharacterHandlers) DeleteCharacterHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(charUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	resp, err := h.charServiceClient.DeleteCharacter(token, charUUID)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// LevelUpHandler godoc
// @Summary Прокачать характеристику персонажа
// @Description Повышает общий уровень на 1 и выбранную характеристику на 1. Характеристика не может превысить 10, уровень - 21.
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body object true "Название характеристики для прокачки" example({"stat": "strength"})
// @Success 200 {object} object "Результат прокачки" example({"message": "Leveled up! New level: 2, New strength: 6"})
// @Failure 400 {object} object "Неверный формат запроса или характеристика уже на максимуме" example({"Error": "Stat is already 10 or character level is 21"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является владельцем" example({"Error": "User is not the owner of this character"})
// @Failure 404 {object} object "Персонаж не найден" example({"Error": "Character not found"})
// @Router /api/character/level-up/{uuid} [post]
func (h *CharacterHandlers) LevelUpHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(charUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		Stat string `json:"stat"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.Stat == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "stat is required"})
	}

	resp, err := h.charServiceClient.LevelUp(token, charUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// GetCharacterInfoHandler godoc
// @Summary Получить информацию о персонаже
// @Description Возвращает полные данные персонажа, включая все характеристики SPECIAL. Доступно только владельцу.
// @Tags Character
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Success 200 {object} object "Полная информация о персонаже" example({"char_uuid": "...", "name": "Vault Dweller", "description": "...", "level": 1, "strength": 5, "perception": 5, "endurance": 5, "charisma": 5, "intelligence": 5, "agility": 5, "luck": 10, "created_at": "2024-06-05T12:00:00Z"})
// @Failure 400 {object} object "Неверный формат UUID" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является владельцем" example({"Error": "User is not the owner of this character"})
// @Failure 404 {object} object "Персонаж не найден" example({"Error": "Character not found"})
// @Router /api/character/info/{uuid} [get]
func (h *CharacterHandlers) GetCharacterInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(charUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	resp, err := h.charServiceClient.GetCharacterInfo(token, charUUID)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Character service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}
