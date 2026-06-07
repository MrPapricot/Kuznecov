package main

import (
	handler "character_service/src/handlers"
	"character_service/src/repository"
	"fmt"
	"log"
	"time"
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
	jwtExpireTime := 30 * 24 * time.Hour
	charPort := utils.ReadEnvU16("CHARACTER_PORT", 8084)

	adapter, err := db_adapter.PostgresConnect(db_adapter.PostgresConnectOptions{
		Host: dbHost, Port: dbPort, UserName: dbUser, UserPassword: dbPassword, DBName: dbName,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer adapter.Close()

	log.Println("Connected to database")

	charRepo := repository.NewCharacterRepository(&adapter, jwtSecret, jwtExpireTime)
	charHandler := handler.NewHandler(charRepo)

	app := fiber.New()

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"Status": "Working"})
	})

	// Character routes
	app.Post("/create", charHandler.CreateCharacterHandler)
	app.Patch("/update-name/:uuid", charHandler.UpdateNameHandler)
	app.Patch("/update-description/:uuid", charHandler.UpdateDescriptionHandler)
	app.Delete("/:uuid", charHandler.DeleteCharacterHandler)
	app.Post("/level-up/:uuid", charHandler.LevelUpHandler)
	app.Get("/info/:uuid", charHandler.GetCharacterInfoHandler)

	app.Listen(fmt.Sprintf("0.0.0.0:%d", charPort))
}
