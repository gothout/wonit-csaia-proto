package chatvolt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"whatsapp-ia-integrator/internal/config"
)

const baseURL = "https://api.chatvolt.ai"

// Client envia consultas para o Chatvolt.
type Client struct {
	httpClient *http.Client
	cfg        config.ChatvoltConfig
}

func NewClient(cfg config.ChatvoltConfig) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg:        cfg,
	}
}

// QueryRequest representa o payload aceito pelo endpoint /agents/{id}/query.
type QueryRequest struct {
	Query          string         `json:"query"`
	Streaming      bool           `json:"streaming"`
	ConversationID string         `json:"conversationId,omitempty"`
	ContactID      string         `json:"contactId,omitempty"`
	Contact        *Contact       `json:"contact,omitempty"`
	VisitorID      string         `json:"visitorId,omitempty"`
	Temperature    *float64       `json:"temperature,omitempty"`
	ModelName      string         `json:"modelName,omitempty"`
	Filters        map[string]any `json:"filters,omitempty"`
	Context        map[string]any `json:"context,omitempty"`
	CallbackURL    string         `json:"callbackURL,omitempty"`
}

// Contact representa dados do remetente enviados para IA.
type Contact struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phoneNumber,omitempty"`
	ExternalID string `json:"userId,omitempty"`
}

// QueryResponse contém o retorno principal do Chatvolt.
type QueryResponse struct {
	Answer         string         `json:"answer"`
	ConversationID string         `json:"conversationId"`
	VisitorID      string         `json:"visitorId"`
	MessageID      string         `json:"messageId"`
	Metadata       map[string]any `json:"metadata"`
}

// Query envia uma mensagem de texto e retorna a resposta da IA.
func (c *Client) Query(ctx context.Context, payload QueryRequest) (*QueryResponse, error) {
	if payload.Query == "" {
		return nil, fmt.Errorf("query text vazio")
	}

	payload.Streaming = false

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal chatvolt payload: %w", err)
	}

	url := fmt.Sprintf("%s/agents/%s/query", baseURL, c.cfg.AgentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new chatvolt request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call chatvolt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var raw map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&raw)
		return nil, fmt.Errorf("chatvolt error: status=%d body=%v", resp.StatusCode, raw)
	}

	var parsed QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode chatvolt response: %w", err)
	}

	return &parsed, nil
}
