package main

import (
	client "MainBackend/src/character_client"
	"MainBackend/src/handlers"
	"MainBackend/src/room_client"
	"MainBackend/src/user_client"
	"context"
	"fmt"
	"log"
	"time"

	"utils"

	"github.com/gofiber/fiber/v3"

	"github.com/segmentio/kafka-go"

	_ "MainBackend/docs"

	swaggo "github.com/gofiber/contrib/v3/swaggo"

	kafka_adapter "MainBackend/src/kafka"
)

// @title Gateway API
// @version 1.0
// @description API Gateway для управления аутентификацией, пользователями и комнатами.
// @host localhost:8000
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey.description Введите токен в формате: Bearer <ваш_токен> или просто <ваш_токен>

type Response struct {
	Message string `json:"message"`
}

func main() {
	ctx := context.Background()

	services_count := 1

	kafka_host := utils.ReadEnv("KAFKA_HOST", "kafka")
	kafka_port := utils.ReadEnvU16("KAFKA_PORT", 9092)
	group_id := "auth_group_2"

	health_topic := utils.ReadEnv("KAFKA_HEALTH_TOPIC", "health")

	brokers := []string{fmt.Sprintf("%s:%d", kafka_host, kafka_port)}

	health_reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   health_topic,
	})

	active_services := 0
	for active_services < services_count {
		message, err := health_reader.FetchMessage(ctx)
		if err == nil {
			fmt.Printf("Message read. Key: %s, Value: %s\n", message.Key, message.Value)
			active_services += 1
			health_reader.CommitMessages(ctx, message)
		} else {
			fmt.Printf("Error reading message: %#v\n", err)
		}
	}

	health_reader.Close()

	requestTopic := utils.ReadEnv("KAFKA_REQUEST_TOPIC", "auth-requests")
	responseTopic := utils.ReadEnv("KAFKA_RESPONSE_TOPIC", "auth-responses")

	kafka_client, err := kafka_adapter.NewKafkaClient(brokers, requestTopic, responseTopic, group_id)

	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafka_client.Close()

	timeout := 20 * time.Second

	auth_handler := handlers.NewAuthHandler(kafka_client, timeout)

	fmt.Println("All services started successfully")

	// -----------
	// User Client
	user_service_timeout := 10 * time.Second
	user_service_url := fmt.Sprintf("http://%s:%d", utils.ReadEnv("USER_HOST", "localhost"), utils.ReadEnvU16("USER_PORT", 8001))

	user_service_client := user_client.NewUserServiceClient(user_service_url, user_service_timeout)
	user_handler := handlers.NewUserHandler(user_service_client)
	// -----------

	// -----------
	// Room Client
	room_service_timeout := 10 * time.Second
	room_service_url := fmt.Sprintf("http://%s:%d", utils.ReadEnv("ROOM_HOST", "localhost"), utils.ReadEnvU16("ROOM_PORT", 8001))

	room_service_client := room_client.NewRoomServiceClient(room_service_url, room_service_timeout)
	room_handler := handlers.NewRoomHandlers(room_service_client)
	// -----------
	// Character Client
	character_service_timeout := 10 * time.Second
	character_service_url := fmt.Sprintf("http://%s:%d", utils.ReadEnv("CHARACTER_HOST", "localhost"), utils.ReadEnvU16("CHARACTER_PORT", 8002))

	character_service_client := client.NewCharacterServiceClient(character_service_url, character_service_timeout)
	character_handler := handlers.NewCharacterHandlers(character_service_client)
	// -----------

	app := fiber.New(fiber.Config{AppName: "API Gateway"})

	app.Get("/health", handlers.HealthHandler)

	app.Get("/swagger/*", swaggo.HandlerDefault)

	app.Post("/api/auth/register", auth_handler.Register)
	app.Post("/api/auth/login", auth_handler.Login)

	// User routes (через HTTP прокси к User Service)
	app.Get("/api/user/info", user_handler.GetUserInfoHandler)
	app.Patch("/api/user/change-username", user_handler.ChangeUsernameHandler)
	app.Patch("/api/user/change-password", user_handler.ChangePasswordHandler)
	app.Post("/api/user/delete", user_handler.DeleteUserHandler)

	// Room routes (через HTTP прокси)
	app.Post("/api/room/create", room_handler.CreateRoomHandler)
	app.Post("/api/room/add-members/:uuid", room_handler.AddMembersHandler)
	app.Get("/api/room/info/:uuid", room_handler.GetRoomInfoHandler)
	app.Post("/api/room/remove-members/:uuid", room_handler.RemoveMembersHandler)
	app.Patch("/api/room/update-name/:uuid", room_handler.UpdateRoomNameHandler)
	app.Patch("/api/room/update-description/:uuid", room_handler.UpdateRoomDescriptionHandler)
	app.Delete("/api/room/:uuid", room_handler.DeleteRoomHandler) // ✅ DELETE метод
	app.Get("/api/room/owned", room_handler.GetOwnedRoomsHandler)
	app.Get("/api/room/joined", room_handler.GetJoinedRoomsHandler)

	app.Post("/api/character/create", character_handler.CreateCharacterHandler)
	app.Patch("/api/character/update-name/:uuid", character_handler.UpdateNameHandler)
	app.Patch("/api/character/update-description/:uuid", character_handler.UpdateDescriptionHandler)
	app.Delete("/api/character/:uuid", character_handler.DeleteCharacterHandler)
	app.Post("/api/character/level-up/:uuid", character_handler.LevelUpHandler)
	app.Get("/api/character/info/:uuid", character_handler.GetCharacterInfoHandler)

	main_host := utils.ReadEnv("MAIN_HOST", "0.0.0.0")
	main_port := utils.ReadEnvU16("MAIN_PORT", 8080)

	err = app.Listen(fmt.Sprintf("%s:%d", main_host, main_port))

	if err != nil {
		fmt.Printf("Error Listening: %#v\n", err)
	}

}
