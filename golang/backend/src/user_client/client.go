package user_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UserServiceClient struct {
	base_url string
	client   *http.Client
}

func NewUserServiceClient(base_url string, timeout time.Duration) *UserServiceClient {
	return &UserServiceClient{
		base_url: base_url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Response представляет ответ от User Service
type Response struct {
	StatusCode int
	Body       []byte
}

// GetUserInfo проксирует GET /user/info
func (client *UserServiceClient) GetUserInfo(token string) (*Response, error) {
	return client.doRequest("GET", "/get_user_info", token, nil)
}

// ChangeUsername проксирует PATCH /user/change-username
func (client *UserServiceClient) ChangeUsername(token string, body any) (*Response, error) {
	json_body, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return client.doRequest("PATCH", "/update_username", token, json_body)
}

// ChangePassword проксирует PATCH /user/change-password
func (client *UserServiceClient) ChangePassword(token string, body any) (*Response, error) {
	json_body, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return client.doRequest("PATCH", "/update_password", token, json_body)
}

// doRequest выполняет HTTP запрос к User Service
func (client *UserServiceClient) doRequest(method, path, token string, body []byte) (*Response, error) {
	url := client.base_url + path

	var body_reader io.Reader
	if body != nil {
		body_reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, body_reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос
	resp, err := client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	resp_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       resp_body,
	}, nil
}
