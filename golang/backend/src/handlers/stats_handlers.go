package handlers

import (
	client "MainBackend/src/stats_client"

	"github.com/gofiber/fiber/v3"
)

type StatsHandlers struct {
	statsServiceClient *client.StatsServiceClient
}

func NewStatsHandlers(client *client.StatsServiceClient) *StatsHandlers {
	return &StatsHandlers{statsServiceClient: client}
}

// GetGlobalStatsHandler godoc
// @Summary Общая статистика сервиса
// @Description Возвращает общее количество пользователей, комнат и персонажей в системе. Доступно без авторизации.
// @Tags Statistics
// @Produce json
// @Success 200 {object} object "Глобальная статистика" example({"total_users": 150, "total_rooms": 42, "total_characters": 310})
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/stats/global [get]
func (h *StatsHandlers) GetGlobalStatsHandler(ctx fiber.Ctx) error {
	resp, err := h.statsServiceClient.GetGlobalStats()
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Stats service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// GetUserStatsHandler godoc
// @Summary Статистика текущего пользователя
// @Description Возвращает дату регистрации, количество созданных персонажей/комнат, количество комнат, где пользователь состоит, а также данные его самого прокачанного персонажа и самой популярной комнаты.
// @Tags Statistics
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object "Статистика пользователя" example({"created_at": "2024-06-01T12:00:00Z", "owned_chars_count": 2, "owned_rooms_count": 1, "joined_rooms_count": 3, "top_character": {"char_uuid": "...", "name": "Bruiser", "level": 5}, "top_room": {"room_uuid": "...", "name": "Main Hub", "member_count": 15}})
// @Failure 401 {object} object "Невалидный токен или пользователь не найден"
// @Failure 500 {object} object "Внутренняя ошибка сервера"
// @Router /api/stats/user [get]
func (h *StatsHandlers) GetUserStatsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	resp, err := h.statsServiceClient.GetUserStats(token)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Stats service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}
