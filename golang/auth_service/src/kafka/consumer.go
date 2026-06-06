package kafka_consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"auth_service/src/service"

	shared "shared/user_data"
)

type KafkaConsumer struct {
	reader      *kafka.Reader
	writer      *kafka.Writer
	authService *service.AuthService
}

const timeout time.Duration = time.Millisecond * 20

func NewKafkaConsumer(
	brokers []string,
	requestTopic, responseTopic, groupID string,
	authService *service.AuthService,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    requestTopic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        responseTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireAll),
	})

	return &KafkaConsumer{
		reader:      reader,
		writer:      writer,
		authService: authService,
	}
}

// Start запускает бесконечный цикл чтения сообщений
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Kafka consumer started")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Context canceled, stopping consumer")
				break
			}
			log.Printf("Error fetching message: %v", err)
			continue
		}

		time.Sleep(timeout)

		// Обрабатываем в горутине
		go c.handleMessage(ctx, msg)
	}
}

// handleMessage обрабатывает одно сообщение
func (c *KafkaConsumer) handleMessage(ctx context.Context, msg kafka.Message) {
	var req shared.AuthRequest
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		log.Printf("Error unmarshaling request: %v", err)
		return
	}

	log.Printf("Processing request: type=%s, correlation_id=%s", req.Type, req.CorrelationID)

	var resp shared.AuthResponse
	resp.CorrelationID = req.CorrelationID

	// Маршрутизация по типу запроса
	switch req.Type {
	case shared.RequestTypeRegister:
		result, err := c.authService.Register(req.Username, req.Email, req.Password)
		if err != nil {
			log.Printf("Register error: %v", err)
			resp = shared.AuthResponse{
				CorrelationID: req.CorrelationID,
				Success:       false,
				Error:         result.Error,
			}
		} else {
			resp = result
			resp.CorrelationID = req.CorrelationID
		}

	case shared.RequestTypeLogin:
		result, err := c.authService.Login(req.Email, req.Password)
		if err != nil {
			log.Printf("Login error: %v", err)
			resp = shared.AuthResponse{
				CorrelationID: req.CorrelationID,
				Success:       false,
				Error:         result.Error,
			}
		} else {
			resp = result
			resp.CorrelationID = req.CorrelationID
		}

	default:
		resp = shared.AuthResponse{
			CorrelationID: req.CorrelationID,
			Success:       false,
			Error:         "Unknown request type",
		}
	}

	// Отправляем ответ в Kafka
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(req.CorrelationID),
		Value: data,
	})
	if err != nil {
		log.Printf("Error sending response: %v", err)
		return
	}

	// Коммитим сообщение ТОЛЬКО после успешной отправки ответа
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("Error committing message: %v", err)
	}

	log.Printf("Response sent for correlation_id=%s", req.CorrelationID)
}

// Close освобождает ресурсы
func (c *KafkaConsumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.writer.Close()
}

// HealthCheck проверяет инициализацию consumer
func (c *KafkaConsumer) HealthCheck() error {
	if c.reader == nil || c.writer == nil {
		return fmt.Errorf("kafka consumer not initialized")
	}
	return nil
}
