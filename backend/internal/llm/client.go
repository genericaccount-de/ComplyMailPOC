package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxErrBodyBytes bounds how much of an error response body we include
// in error messages, to avoid logging large payloads.
const maxErrBodyBytes = 512

// Client abstracts an OpenAI-compatible chat backend so handlers and
// domain methods can depend on an interface (and be faked in tests).
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// HTTPClient is a Client that talks to an OpenAI-compatible HTTP endpoint.
type HTTPClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// compile-time check that HTTPClient satisfies Client.
var _ Client = (*HTTPClient)(nil)

// New builds an HTTPClient from cfg, applying defaults for empty fields.
func New(cfg Config) *HTTPClient {
	cfg = cfg.withDefaults()
	return &HTTPClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		http:    &http.Client{Timeout: cfg.Timeout},
	}
}

// Model returns the configured model identifier.
func (c *HTTPClient) Model() string { return c.model }

// APIError describes a non-2xx response from the endpoint.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm: unexpected status %d: %s", e.StatusCode, e.Body)
}

// Chat sends a chat-completions request. If req.Model is empty, the
// client's configured model is used.
func (c *HTTPClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.model
	}

	body, err := json.Marshal(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBodyBytes))
		return ChatResponse{}, &APIError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(snippet)),
		}
	}

	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ChatResponse{}, fmt.Errorf("llm: decode response: %w", err)
	}
	return out, nil
}
