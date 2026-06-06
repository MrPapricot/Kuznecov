package main

import (
	"fmt"
	"time"
	"user_service/src/handlers"
	"user_service/src/repository"
	"utils"

	db_adapter "db_adapter/src"

	"github.com/gofiber/fiber/v3"
)

func main() {
	user_port := utils.ReadEnvU16("USER_PORT", 8001)

	db_host := utils.ReadEnv("DB_HOST", "kafka")
	db_port := utils.ReadEnvU16("DB_PORT", 5432)

	db_user := utils.ReadEnv("DB_USER", "postgres")
	db_user_password := utils.ReadEnv("DB_USER_PASSWORD", "1234")

	db_name := utils.ReadEnv("DB_NAME", "Kuznetsov")

	jwt_secret := utils.ReadEnv("JWT_SECRET", "ajkghjkag(ajkng8934_)(&Yag")
	token_ttl := 30 * 24 * time.Hour // 1 месяц

	connect_options := db_adapter.PostgresConnectOptions{
		UserName:     db_user,
		UserPassword: db_user_password,
		Host:         db_host,
		Port:         db_port,
		DBName:       db_name,
	}
	adapter, err := db_adapter.PostgresConnect(connect_options)

	if err == nil {
		fmt.Printf("Connected Successfully with options: %#v\n", connect_options)
	} else {
		fmt.Printf("Error connecting to database with options %#v\nError is %#v\n", connect_options, err)
		panic("Error connecting to DB")
	}

	defer adapter.Close()

	repository := repository.NewUserRepository(&adapter, jwt_secret, token_ttl)

	handler := handlers.NewHandler(repository)

	app := fiber.New(fiber.Config{AppName: "User Service"})

	app.Get("/", func(ctx fiber.Ctx) error {
		return ctx.SendString("Hello")
	})

	app.Get("/get_user_info", handler.GetUserInfoHandler)

	app.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"Status": "Working"})
	})

	app.Patch("/update_username", handler.ChangeUsernameHandler)

	app.Patch("/update_password", handler.ChangePasswordHandler)

	app.Post("/delete", handler.DeleteUserHandler)

	fmt.Println("Start serving")

	app.Listen(fmt.Sprintf("%s:%d", "0.0.0.0", user_port))
}
