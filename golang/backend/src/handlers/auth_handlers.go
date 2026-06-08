package handlers

import (
	shared "shared/user_data"

	kafka_adapter "MainBackend/src/kafka"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthHandler struct {
	kafkaClient *kafka_adapter.KafkaClient
	timeout     time.Duration
}

func NewAuthHandler(kafkaClient *kafka_adapter.KafkaClient, timeout time.Duration) *AuthHandler {
	return &AuthHandler{
		kafkaClient: kafkaClient,
		timeout:     timeout,
	}
}

// Login godoc
// @Summary Вход в систему (Логин)
// @Description Аутентифицирует пользователя по email и паролю, возвращает JWT токен
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body handlers.LoginRequest true "Данные для входа"
// @Success 200 {object} object "Успешный вход" example({"token": "eyJhbG...", "expires_at": 1717600000, "user_id": "550e8400-..."})
// @Failure 401 {object} object "Неверные учетные данные" example({"Error": "Invalid credentials"})
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
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

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает аккаунт пользователя и возвращает JWT токен
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body handlers.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} object "Успешная регистрация" example({"token": "eyJhbG...", "expires_at": 1717600000, "user_id": "550e8400-..."})
// @Failure 400 {object} object "Ошибка валидации" example({"Error": "Email already registered"})
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c fiber.Ctx) error {
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
