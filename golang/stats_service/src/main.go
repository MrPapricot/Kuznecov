package main

import (
	"fmt"
	"log"
	handler "stats_service/src/handlers"
	"stats_service/src/repository"
	"utils"

	db_adapter "db_adapter/src"

	"github.com/gofiber/fiber/v3"
)

func main() {
	dbHost := utils.ReadEnv("DB_HOST", "postgres")
	dbPort := utils.ReadEnvU16("DB_PORT", 5432)
	dbUser := utils.ReadEnv("DB_USER", "auth_user")
	dbPassword := utils.ReadEnv("DB_USER_PASSWORD", "auth_pass")
	dbName := utils.ReadEnv("DB_NAME", "auth_db")
	jwtSecret := utils.ReadEnv("JWT_SECRET", "super-secret-key-change-in-production")
	statsPort := utils.ReadEnvU16("STATS_PORT", 8085)

	adapter, err := db_adapter.PostgresConnect(db_adapter.PostgresConnectOptions{
		Host: dbHost, Port: dbPort, UserName: dbUser, UserPassword: dbPassword, DBName: dbName,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer adapter.Close()

	statsRepo := repository.NewStatsRepository(&adapter, jwtSecret)
	statsHandler := handler.NewHandler(statsRepo)

	app := fiber.New()

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/global", statsHandler.GetGlobalStatsHandler)
	app.Get("/user", statsHandler.GetUserStatsHandler)

	fmt.Println("Start serving")

	app.Listen(fmt.Sprintf("%s:%d", "0.0.0.0", statsPort))
}
