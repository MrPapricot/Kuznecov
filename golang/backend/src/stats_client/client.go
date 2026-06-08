package stats_client

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	StatusCode int
	Body       []byte
}

type StatsServiceClient struct {
	baseURL string
	client  *http.Client
}

func NewStatsServiceClient(baseURL string, timeout time.Duration) *StatsServiceClient {
	return &StatsServiceClient{baseURL: baseURL, client: &http.Client{Timeout: timeout}}
}

func (c *StatsServiceClient) GetGlobalStats() (*Response, error) {
	return c.doRequest("GET", "/global", "", nil)
}

func (c *StatsServiceClient) GetUserStats(token string) (*Response, error) {
	return c.doRequest("GET", "/user", token, nil)
}

func (c *StatsServiceClient) doRequest(method, path, token string, body []byte) (*Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
}
