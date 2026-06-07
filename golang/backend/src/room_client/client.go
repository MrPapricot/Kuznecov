package room_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RoomServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewRoomServiceClient(baseURL string, timeout time.Duration) *RoomServiceClient {
	return &RoomServiceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type Response struct {
	StatusCode int
	Body       []byte
}

// doRequest выполняет HTTP запрос к Room Service
func (c *RoomServiceClient) doRequest(method, path, token string, body []byte) (*Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

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

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}, nil
}

func (c *RoomServiceClient) CreateRoom(token string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("POST", "/create", token, jsonBody)
}

func (c *RoomServiceClient) AddMembers(token, roomUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("POST", fmt.Sprintf("/add-members/%s", roomUUID), token, jsonBody)
}

func (c *RoomServiceClient) GetRoomInfo(token, roomUUID string) (*Response, error) {
	return c.doRequest("GET", fmt.Sprintf("/info/%s", roomUUID), token, nil)
}

func (c *RoomServiceClient) RemoveMembers(token, roomUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("POST", fmt.Sprintf("/remove-members/%s", roomUUID), token, jsonBody)
}

func (c *RoomServiceClient) UpdateRoomName(token, roomUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("PATCH", fmt.Sprintf("/update-name/%s", roomUUID), token, jsonBody)
}

func (c *RoomServiceClient) UpdateRoomDescription(token, roomUUID string, body any) (*Response, error) {
	jsonBody, _ := json.Marshal(body)
	return c.doRequest("PATCH", fmt.Sprintf("/update-description/%s", roomUUID), token, jsonBody)
}

func (c *RoomServiceClient) DeleteRoom(token, roomUUID string) (*Response, error) {
	// Тело nil, так как uuid уже в пути, а метод DELETE
	return c.doRequest("DELETE", fmt.Sprintf("/%s", roomUUID), token, nil)
}

func (c *RoomServiceClient) GetOwnedRooms(token string) (*Response, error) {
	return c.doRequest("GET", "/owned", token, nil)
}

func (c *RoomServiceClient) GetJoinedRooms(token string) (*Response, error) {
	return c.doRequest("GET", "/joined", token, nil)
}
