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
// @Description Доступно без авторизации.
// @Tags Statistics
// @Produce json
// @Success 200 {object} handlers.GlobalStatsResponse "Глобальная статистика"
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
// @Description Поля top_character и top_room будут null, если у пользователя нет персонажей или комнат.
// @Tags Statistics
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handlers.UserStatsResponse "Статистика пользователя"
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
