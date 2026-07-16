package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Sensitivity levels returned by ClassifySensitivity.
const (
	SensitivityLow    = "LOW"
	SensitivityMedium = "MEDIUM"
	SensitivityHigh   = "HIGH"
)

// StyleSuggestion is a single style/compliance recommendation.
type StyleSuggestion struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// SensitivityResult is the outcome of a sensitivity classification.
type SensitivityResult struct {
	Level   string   `json:"level"`
	Reasons []string `json:"reasons"`
}

// jsonResponseFormat requests JSON mode from the endpoint.
var jsonResponseFormat = &ResponseFormat{Type: "json_object"}

// AnalyzeStyle reviews an email against the given style guide and returns
// zero or more suggestions.
func (c *HTTPClient) AnalyzeStyle(ctx context.Context, styleGuide string, email Email) ([]StyleSuggestion, error) {
	resp, err := c.Chat(ctx, ChatRequest{
		Messages:    buildStyleMessages(styleGuide, email),
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}

	content, err := firstContent(resp)
	if err != nil {
		return nil, err
	}

	// The style prompt asks for a bare JSON array. Some models wrap it in an
	// object; support both shapes.
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "{") {
		var wrapper struct {
			Suggestions []StyleSuggestion `json:"suggestions"`
		}
		if err := json.Unmarshal([]byte(trimmed), &wrapper); err != nil {
			return nil, fmt.Errorf("llm: parse style object: %w (content=%q)", err, snip(content))
		}
		return wrapper.Suggestions, nil
	}

	var suggestions []StyleSuggestion
	if err := json.Unmarshal([]byte(trimmed), &suggestions); err != nil {
		return nil, fmt.Errorf("llm: parse style array: %w (content=%q)", err, snip(content))
	}
	return suggestions, nil
}

// ClassifySensitivity classifies the sensitivity of an email's content.
func (c *HTTPClient) ClassifySensitivity(ctx context.Context, email Email) (SensitivityResult, error) {
	resp, err := c.Chat(ctx, ChatRequest{
		Messages:       buildSensitivityMessages(email),
		Temperature:    0,
		ResponseFormat: jsonResponseFormat,
	})
	if err != nil {
		return SensitivityResult{}, err
	}

	content, err := firstContent(resp)
	if err != nil {
		return SensitivityResult{}, err
	}

	var result SensitivityResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &result); err != nil {
		return SensitivityResult{}, fmt.Errorf("llm: parse sensitivity: %w (content=%q)", err, snip(content))
	}

	result.Level = strings.ToUpper(strings.TrimSpace(result.Level))
	switch result.Level {
	case SensitivityLow, SensitivityMedium, SensitivityHigh:
		return result, nil
	default:
		return SensitivityResult{}, fmt.Errorf("llm: invalid sensitivity level %q", result.Level)
	}
}

// firstContent extracts the assistant message content from the first choice.
func firstContent(resp ChatResponse) (string, error) {
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("llm: response contained no choices")
	}
	return resp.Choices[0].Message.Content, nil
}

// snip truncates s for inclusion in error messages.
func snip(s string) string {
	const max = 200
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
