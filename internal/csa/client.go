package csa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"whatsapp-ia-integrator/internal/config"
)

const (
	defaultCSAURL     = "https://csa.wonit.net.br"
	defaultProduct    = "whatsapp"
	defaultProvider   = "gupshup"
	defaultInstanceID = "3f9e541b-90b1-4052-abef-f1835a43e470" // AJUSTA para tua instância
	defaultSenderName = "Wonit Tecnologia"
)

type Client struct {
	httpClient *http.Client
	cfg        config.CSAConfig
}

func NewClient(cfg config.CSAConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cfg:        cfg,
	}
}

type SendMessageRequest struct {
	Address     string `json:"address,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Destination string `json:"destination"`
	Filename    string `json:"filename,omitempty"`
	InstanceID  string `json:"instanceId"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	Name        string `json:"name,omitempty"`
	Preview     bool   `json:"preview"`
	Product     string `json:"product"`
	Provider    string `json:"provider"`
	Text        string `json:"text,omitempty"`
	Type        string `json:"type"` // "text", "image", "document", etc.
	URL         string `json:"url,omitempty"`
}

type SendMessageResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
}

func (c *Client) SendMessage(ctx context.Context, payload *SendMessageRequest) (*SendMessageResponse, error) {
	// Defaults
	if payload.InstanceID == "" {
		payload.InstanceID = defaultInstanceID
	}
	if payload.Product == "" {
		payload.Product = defaultProduct
	}
	if payload.Provider == "" {
		payload.Provider = defaultProvider
	}
	if payload.Name == "" {
		payload.Name = defaultSenderName
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal csa payload: %w", err)
	}

	baseURL := c.cfg.URL
	if baseURL == "" {
		baseURL = defaultCSAURL
	}

	url := fmt.Sprintf("%s/api/integration/whatsapp/%s/send", baseURL, c.cfg.WebhookID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new csa request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", c.cfg.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call csa: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var raw map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		return nil, fmt.Errorf("csa error: status=%d body=%v", resp.StatusCode, raw)
	}

	var data SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode csa response: %w", err)
	}

	return &data, nil
}
