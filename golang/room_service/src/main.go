package main

import (
	"fmt"
	"log"
	handler "room_service/src/handlers"
	"room_service/src/repository"
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
	roomPort := utils.ReadEnvU16("ROOM_PORT", 8002)

	adapter, err := db_adapter.PostgresConnect(db_adapter.PostgresConnectOptions{
		Host: dbHost, Port: dbPort, UserName: dbUser, UserPassword: dbPassword, DBName: dbName,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer adapter.Close()

	log.Println("Connected to database")

	roomRepo := repository.NewRoomRepository(&adapter, jwtSecret, jwtExpireTime)
	roomHandler := handler.NewHandler(roomRepo)

	app := fiber.New()

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"Status": "Working"})
	})

	// ✅ RESTful роуты без префиксов, uuid в path
	app.Post("/create", roomHandler.CreateRoomHandler)
	app.Post("/add-members/:uuid", roomHandler.AddMembersHandler)
	app.Get("/info/:uuid", roomHandler.GetRoomInfoHandler)
	app.Post("/remove-members/:uuid", roomHandler.RemoveMembersHandler)
	app.Patch("/update-name/:uuid", roomHandler.UpdateRoomNameHandler)
	app.Patch("/update-description/:uuid", roomHandler.UpdateRoomDescriptionHandler)
	app.Delete("/:uuid", roomHandler.DeleteRoomHandler) // ✅ DELETE метод
	app.Get("/owned", roomHandler.GetOwnedRoomsHandler)
	app.Get("/joined", roomHandler.GetJoinedRoomsHandler)

	app.Listen(fmt.Sprintf("0.0.0.0:%d", roomPort))
}
