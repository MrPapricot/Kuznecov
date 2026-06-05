package main

import (
	"MainBackend/src/handlers"
	"context"
	"fmt"
	"log"
	"time"

	"utils"

	"github.com/gofiber/fiber/v3"

	"github.com/segmentio/kafka-go"

	kafka_adapter "MainBackend/src/kafka"
)

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

	handler := handlers.NewHandler(kafka_client, timeout)

	fmt.Println("All services started successfully")

	app := fiber.New(fiber.Config{AppName: "API Gateway"})

	app.Get("/health", handlers.HealthHandler)

	app.Post("/api/auth/register", handler.Register)
	app.Post("/api/auth/login", handler.Login)

	main_host := utils.ReadEnv("MAIN_HOST", "0.0.0.0")
	main_port := utils.ReadEnvU16("MAIN_PORT", 8080)

	err = app.Listen(fmt.Sprintf("%s:%d", main_host, main_port))

	if err != nil {
		fmt.Printf("Error Listening: %#v\n", err)
	}

}
