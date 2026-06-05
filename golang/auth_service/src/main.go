package main

import (
	kafka_consumer "auth_service/src/kafka"
	"auth_service/src/repository"
	"auth_service/src/service"
	"context"
	db_adapter "db_adapter/src"
	"fmt"
	"time"
	"utils"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	kafka_host := utils.ReadEnv("KAFKA_HOST", "kafka")
	kafka_port := utils.ReadEnvU16("KAFKA_PORT", 9092)

	kafka_broker := []string{fmt.Sprintf("%s:%d", kafka_host, kafka_port)}

	health_topic := utils.ReadEnv("KAFKA_HEALTH_TOPIC", "health")
	request_topic := utils.ReadEnv("KAFKA_REQUEST_TOPIC", "auth-requests")
	response_topic := utils.ReadEnv("KAFKA_RESPONSE_TOPIC", "auth-responses")
	group_id := "auth_group"

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

	health_writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{fmt.Sprintf("%s:%d", kafka_host, kafka_port)},
		Topic:        health_topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: int(kafka.RequireAll),
		WriteTimeout: time.Second * 10,
	})

	err = health_writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("AUTH"),
		Value: []byte("Active"),
	})

	if err != nil {
		fmt.Printf("Error sending health report %#v\n", err)
	} else {
		fmt.Println("Health report sent successfully")
	}

	user_repo := repository.NewUserRepository(&adapter)
	auth_service := service.NewAuthService(user_repo, jwt_secret, token_ttl)

	// Инициализация Kafka consumer
	consumer := kafka_consumer.NewKafkaConsumer(
		kafka_broker,
		request_topic,
		response_topic,
		group_id,
		auth_service,
	)
	defer consumer.Close()

	consumer.Start(ctx)
}
