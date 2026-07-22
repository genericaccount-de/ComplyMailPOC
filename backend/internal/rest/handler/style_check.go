package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/genericaccount-de/comply-mail-poc/backend/internal/llm"
)

// defaultStyleGuide is a minimal built-in style guide used when no
// customer-specific guide is configured. Kept short for the POC.
const defaultStyleGuide = `- Use a professional greeting (e.g. "Hi <name>," or "Dear <name>,").
- Avoid slang and overly casual language.
- Include a closing line and a sign-off (e.g. "Best regards,").
- Do not use ALL CAPS or excessive exclamation marks.
- Be concise and courteous.`

// StyleAnalyzer is the subset of the LLM client the compose handler needs.
// Defining it here (rather than depending on *llm.HTTPClient) lets tests
// supply a fake implementation.
type StyleAnalyzer interface {
	AnalyzeStyle(ctx context.Context, styleGuide string, email llm.Email) ([]llm.StyleSuggestion, error)
}

// StyleCheck handles POST /check-style-email: it reviews a draft email and
// returns style/compliance suggestions for display in the Outlook add-in.
type StyleCheck struct {
	analyzer   StyleAnalyzer
	styleGuide string
}

// NewStyleCheck builds a StyleCheck handler. If styleGuide is empty, a built-in
// default guide is used.
func NewStyleCheck(analyzer StyleAnalyzer, styleGuide string) *StyleCheck {
	if strings.TrimSpace(styleGuide) == "" {
		styleGuide = defaultStyleGuide
	}
	return &StyleCheck{analyzer: analyzer, styleGuide: styleGuide}
}

// composeRequest is the JSON body sent by the Outlook add-in.
type styleCheckRequest struct {
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	Recipients []string `json:"recipients"`
}

// composeResponse is returned to the add-in. Suggestions is always non-nil so
// clients can rely on receiving a JSON array.
type styleCheckResponse struct {
	Suggestions []llm.StyleSuggestion `json:"suggestions"`
}

// ServeHTTP implements http.Handler.
func (h *StyleCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req styleCheckRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Subject) == "" && strings.TrimSpace(req.Body) == "" {
		writeError(w, http.StatusBadRequest, "subject or body is required")
		return
	}

	suggestions, err := h.analyzer.AnalyzeStyle(r.Context(), h.styleGuide, llm.Email{
		Subject:    req.Subject,
		Body:       req.Body,
		Recipients: req.Recipients,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "style analysis failed")
		return
	}

	if suggestions == nil {
		suggestions = []llm.StyleSuggestion{}
	}
	writeJSON(w, http.StatusOK, styleCheckResponse{Suggestions: suggestions})
}
