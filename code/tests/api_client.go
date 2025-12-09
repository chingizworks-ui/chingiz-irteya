package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"stockpilot/internal/config"
	"stockpilot/internal/handler"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewAPIClient(cfg config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    normalizeBaseURL(cfg.ListenAddr),
	}
}

func (c *Client) RegisterUser(req handler.RegisterUserRequest) (*http.Response, error) {
	return c.post("/api/v1/users/register", req)
}

func (c *Client) CreateProduct(req handler.CreateProductRequest) (*http.Response, error) {
	return c.post("/api/v1/products", req)
}

func (c *Client) CreateOrder(req handler.CreateOrderRequest) (*http.Response, error) {
	return c.post("/api/v1/orders", req)
}

func (c *Client) GetProduct(id string) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, strings.TrimLeft(id, "/"))
	httpReq, err := http.NewRequest(http.MethodGet, fullURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	return resp, nil
}

func (c *Client) post(path string, body any) (*http.Response, error) {
	fullURL := c.baseURL + path
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	return resp, nil
}

func normalizeBaseURL(addr string) string {
	host := strings.TrimSpace(addr)
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimSuffix(host, "/")
	}
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host
	}
	if strings.HasPrefix(host, "0.0.0.0") {
		host = strings.Replace(host, "0.0.0.0", "127.0.0.1", 1)
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return strings.TrimSuffix(host, "/")
}
