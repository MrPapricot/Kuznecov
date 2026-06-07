package handler

import (
	"character_service/src/repository"
	"errors"
	"fmt"
	"time"

	db_adapter "db_adapter/src"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	repository repository.CharacterRepository
}

func NewHandler(repository repository.CharacterRepository) Handler {
	return Handler{repository: repository}
}

// CreateCharacterHandler POST /create
func (h *Handler) CreateCharacterHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Level        *int16 `json:"level"` // Указатель, чтобы отличить 0 от непереданного значения
		Strength     *int16 `json:"strength"`
		Perception   *int16 `json:"perception"`
		Endurance    *int16 `json:"endurance"`
		Charisma     *int16 `json:"charisma"`
		Intelligence *int16 `json:"intelligence"`
		Agility      *int16 `json:"agility"`
		Luck         *int16 `json:"luck"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		fmt.Printf("⚠️ Character Create Bind Error: %v\n", err)
		fmt.Printf("📥 Raw Body: %s\n", string(ctx.Body()))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.Name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "name is required"})
	}

	level := int16(1)
	if req.Level != nil {
		level = *req.Level
	}

	stats := db_adapter.CreateCharacterBody{
		Strength:     getValue(req.Strength, 0),
		Perception:   getValue(req.Perception, 0),
		Endurance:    getValue(req.Endurance, 0),
		Charisma:     getValue(req.Charisma, 0),
		Intelligence: getValue(req.Intelligence, 0),
		Agility:      getValue(req.Agility, 0),
		Luck:         getValue(req.Luck, 0),
	}

	charUUID, err := h.repository.CreateCharacter(token, req.Name, req.Description, level, stats)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"char_uuid": charUUID.String()})
}

// UpdateNameHandler PATCH /update-name/:uuid
func (h *Handler) UpdateNameHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
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

	oldName, newName, err := h.repository.UpdateCharacterName(token, charUUID, req.NewName)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"old_name": oldName, "new_name": newName})
}

// UpdateDescriptionHandler PATCH /update-description/:uuid
func (h *Handler) UpdateDescriptionHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewDescription string `json:"new_description"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	oldDesc, newDesc, err := h.repository.UpdateCharacterDescription(token, charUUID, req.NewDescription)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"old_description": oldDesc, "new_description": newDesc})
}

// DeleteCharacterHandler DELETE /:uuid
func (h *Handler) DeleteCharacterHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	err = h.repository.DeleteCharacter(token, charUUID)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"message": "Character deleted successfully"})
}

// LevelUpHandler POST /level-up/:uuid
func (h *Handler) LevelUpHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
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

	newLevel, newStatVal, err := h.repository.LevelUpCharacter(token, charUUID, req.Stat)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{
		"message": fmt.Sprintf("Leveled up! New level: %d, New %s: %d", newLevel, req.Stat, newStatVal),
	})
}

// GetCharacterInfoHandler GET /info/:uuid
func (h *Handler) GetCharacterInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	charUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	char, err := h.repository.GetCharacterInfo(token, charUUID)
	if err != nil {
		return h.handleError(ctx, err)
	}

	desc := ""
	if char.CharacterDescription.Valid {
		desc = char.CharacterDescription.String
	}

	return ctx.JSON(fiber.Map{
		"char_uuid":    char.CharUUID.String(),
		"name":         char.Name,
		"description":  desc,
		"level":        char.Level,
		"strength":     char.Strength,
		"perception":   char.Perception,
		"endurance":    char.Endurance,
		"charisma":     char.Charisma,
		"intelligence": char.Intelligence,
		"agility":      char.Agility,
		"luck":         char.Luck,
		"created_at":   char.CreatedAt.UTC().Format(time.RFC3339),
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
	case errors.Is(err, repository.NotCharacterOwner):
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"Error": "User is not the owner of this character"})
	case errors.Is(err, repository.CharacterNotFound):
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"Error": "Character not found"})
	case errors.Is(err, repository.InvalidStatSum):
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid stat points sum. Must equal 40 + (level - 1)"})
	case errors.Is(err, repository.InvalidStatValue):
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Each stat must be between 0 and 10"})
	case errors.Is(err, repository.InvalidLevel):
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Level must be between 1 and 21"})
	case errors.Is(err, repository.InvalidStatName):
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid stat name. Use: strength, perception, endurance, charisma, intelligence, agility, luck"})
	case errors.Is(err, repository.StatMaxedOrLevelMaxed):
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Stat is already 10 or character level is 21"})
	case errors.Is(err, repository.DatabaseError):
		fmt.Printf("Database error: %#v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	default:
		fmt.Printf("Base error: %#v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	}
}

func getValue(ptr *int16, def int16) int16 {
	if ptr == nil {
		return def
	}
	return *ptr
}
