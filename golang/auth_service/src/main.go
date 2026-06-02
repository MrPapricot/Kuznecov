package main

import (
	"auth_service/src/utils"
	"context"
	db_adapter "db_adapter/src"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	kafka_host := utils.ReadEnv("KAFKA_HOST", "kafka")
	kafka_port := utils.ReadEnvU16("KAFKA_PORT", 9092)

	health_topic := utils.ReadEnv("KAFKA_HEALTH_TOPIC", "health")

	db_host := utils.ReadEnv("DB_HOST", "kafka")
	db_port := utils.ReadEnvU16("DB_PORT", 5432)

	db_user := utils.ReadEnv("DB_SUPERUSER", "postgres")
	db_user_password := utils.ReadEnv("DB_SUPERUSER_PASSWORD", "1234")

	db_name := utils.ReadEnv("DB_NAME", "Kuznetsov")

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
		fmt.Println("Error connecting to Database", err)
		panic("Error connecting to DB")
	}
	_ = adapter

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{fmt.Sprintf("%s:%d", kafka_host, kafka_port)},
		Topic:        health_topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: int(kafka.RequireAll),
		WriteTimeout: time.Second * 10,
	})

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("AUTH"),
		Value: []byte("Active"),
	})

	if err != nil {
		fmt.Printf("Error sending message %#v", err)
	}

	fmt.Println("Hello from auth service")
	fmt.Println("Hello again")
	for true {

	}
}
