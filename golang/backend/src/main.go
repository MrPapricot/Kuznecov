package main

import (
	"MainBackend/src/handlers"
	"context"
	"fmt"

	"utils"

	"github.com/gofiber/fiber/v3"

	"github.com/segmentio/kafka-go"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	ctx := context.Background()

	kafka_host := utils.ReadEnv("KAFKA_HOST", "kafka")
	kafka_port := utils.ReadEnvU16("KAFKA_PORT", 9092)

	health_topic := utils.ReadEnv("KAFKA_HEALTH_TOPIC", "health")

	health_reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{fmt.Sprintf("%s:%d", kafka_host, kafka_port)},
		Topic:   health_topic,
	})

	message, err := health_reader.ReadMessage(ctx)

	if err == nil {
		fmt.Printf("Message read. Key: %s, Value: %s\n", message.Key, message.Value)
	} else {
		fmt.Printf("Error reading message: %#v\n", err)
	}

	app := fiber.New()

	app.Get("/", handlers.HelloHandler)

	app.Get("/health", handlers.HealthHandler)

	app.Get("/hello", func(c fiber.Ctx) error {
		response := Response{
			Message: "Hello",
		}
		return c.JSON(response)
	})

	app.Listen(":8000")
}
