package kafka_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	shared "shared/auth"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	writer  *kafka.Writer
	reader  *kafka.Reader
	pending map[string]chan shared.AuthResponse
	mu      sync.RWMutex
}

func NewKafkaClient(brokers []string, requestTopic, responseTopic, groupID string) (*KafkaClient, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        requestTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireAll),
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    responseTopic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	client := &KafkaClient{
		writer:  writer,
		reader:  reader,
		pending: make(map[string]chan shared.AuthResponse),
	}

	// Запускаем фоновое чтение ответов
	go client.readResponses()

	return client, nil
}

// SendRequest отправляет запрос и ждет ответа (с таймаутом)
func (c *KafkaClient) SendRequest(ctx context.Context, req shared.AuthRequest, timeout time.Duration) (shared.AuthResponse, error) {
	// Создаем канал для ожидания ответа
	ch := make(chan shared.AuthResponse, 1)

	// Регистрируем канал в мапе
	c.mu.Lock()
	c.pending[req.CorrelationID] = ch
	c.mu.Unlock()

	// Отправляем запрос в Kafka
	data, err := json.Marshal(req)
	if err != nil {
		c.cleanup(req.CorrelationID)
		return shared.AuthResponse{}, fmt.Errorf("marshal error: %w", err)
	}

	err = c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(req.CorrelationID),
		Value: data,
	})
	if err != nil {
		c.cleanup(req.CorrelationID)
		return shared.AuthResponse{}, fmt.Errorf("kafka write error: %w", err)
	}

	// Ждем ответ с таймаутом
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		c.cleanup(req.CorrelationID)
		return shared.AuthResponse{}, fmt.Errorf("timeout waiting for response")
	case <-ctx.Done():
		c.cleanup(req.CorrelationID)
		return shared.AuthResponse{}, ctx.Err()
	}
}

// readResponses читает ответы из Kafka и направляет их в нужные каналы
func (c *KafkaClient) readResponses() {
	for {
		msg, err := c.reader.FetchMessage(context.Background())
		if err != nil {
			log.Printf("Error reading response: %v", err)
			continue
		}

		var resp shared.AuthResponse
		if err := json.Unmarshal(msg.Value, &resp); err != nil {
			log.Printf("Error unmarshaling response: %v", err)
			continue
		}

		// Находим канал для этого correlation_id
		c.mu.RLock()
		ch, exists := c.pending[resp.CorrelationID]
		c.mu.RUnlock()

		if exists {
			ch <- resp
		} else {
			log.Printf("No pending request for correlation_id: %s", resp.CorrelationID)
		}

		// Коммитим сообщение
		if err := c.reader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("Error committing message: %v", err)
		}
	}
}

func (c *KafkaClient) cleanup(correlationID string) {
	c.mu.Lock()
	delete(c.pending, correlationID)
	c.mu.Unlock()
}

func (c *KafkaClient) Close() error {
	if err := c.writer.Close(); err != nil {
		return err
	}
	return c.reader.Close()
}
