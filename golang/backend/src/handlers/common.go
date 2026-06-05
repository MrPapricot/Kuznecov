package handlers

import (
	kafka_adapter "MainBackend/src/kafka"
	"time"
)

type Handler struct {
	kafkaClient *kafka_adapter.KafkaClient
	timeout     time.Duration
}

func NewHandler(kafkaClient *kafka_adapter.KafkaClient, timeout time.Duration) *Handler {
	return &Handler{
		kafkaClient: kafkaClient,
		timeout:     timeout,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
