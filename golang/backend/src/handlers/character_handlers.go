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
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CreateCharacterRequest true "Данные персонажа"
// @Success 201 {object} handlers.CreateCharacterResponse "UUID созданного персонажа"
// @Failure 400 {object} handlers.ErrorResponse "Ошибка валидации суммы очков"
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
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body handlers.UpdateCharacterNameRequest true "Новое имя"
// @Success 200 {object} handlers.UpdateCharacterNameResponse "Старое и новое имя"
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
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body handlers.UpdateCharacterDescriptionRequest true "Новое описание"
// @Success 200 {object} handlers.UpdateCharacterDescriptionResponse "Старое и новое описание"
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
// @Tags Character
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Success 200 {object} handlers.MessageResponse "Сообщение об успехе"
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
// @Tags Character
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Param request body handlers.LevelUpCharacterRequest true "Название характеристики"
// @Success 200 {object} handlers.MessageResponse "Результат прокачки"
// @Failure 400 {object} handlers.ErrorResponse "Характеристика или уровень на максимуме"
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
// @Tags Character
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID персонажа"
// @Success 200 {object} handlers.CharacterInfoResponse "Полная информация о персонаже"
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
