package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CharacterServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewCharacterServiceClient(baseURL string, timeout time.Duration) *CharacterServiceClient {
	return &CharacterServiceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (c *CharacterServiceClient) doRequest(method, path, token string, body []byte) (*Response, error) {
	url := c.baseURL + path
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	fmt.Println("Character accessed")

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
}

func (c *CharacterServiceClient) CreateCharacter(token string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("POST", "/create", token, jsonBody)
}

func (c *CharacterServiceClient) UpdateName(token, charUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("PATCH", fmt.Sprintf("/update-name/%s", charUUID), token, jsonBody)
}

func (c *CharacterServiceClient) UpdateDescription(token, charUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("PATCH", fmt.Sprintf("/update-description/%s", charUUID), token, jsonBody)
}

func (c *CharacterServiceClient) DeleteCharacter(token, charUUID string) (*Response, error) {
	return c.doRequest("DELETE", fmt.Sprintf("/%s", charUUID), token, nil)
}

func (c *CharacterServiceClient) LevelUp(token, charUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("POST", fmt.Sprintf("/level-up/%s", charUUID), token, jsonBody)
}

func (c *CharacterServiceClient) GetCharacterInfo(token, charUUID string) (*Response, error) {
	return c.doRequest("GET", fmt.Sprintf("/info/%s", charUUID), token, nil)
}
