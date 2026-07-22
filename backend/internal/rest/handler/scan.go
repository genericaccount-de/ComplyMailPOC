package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/genericaccount-de/comply-mail-poc/backend/internal/llm"
)

// Actions returned in a ScanResponse, indicating what the SMTP proxy should
// do with the message.
const (
	ActionPass     = "pass"
	ActionFlag     = "flag"
	ActionRedirect = "redirect"
)

// SensitivityClassifier is the subset of the LLM client the scan handler
// needs. Defining it here lets tests supply a fake implementation.
type SensitivityClassifier interface {
	ClassifySensitivity(ctx context.Context, email llm.Email) (llm.SensitivityResult, error)
}

// Scan handles POST /scan-outbound-email: it classifies the sensitivity of an
// outbound email and tells the SMTP proxy whether to pass, flag, or redirect.
type Scan struct {
	classifier SensitivityClassifier
}

// NewScan builds a Scan handler.
func NewScan(classifier SensitivityClassifier) *Scan {
	return &Scan{classifier: classifier}
}

// scanRequest is the JSON body sent by the SMTP proxy for each outbound email.
type scanRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

// scanResponse is returned to the SMTP proxy. Flags is always non-nil so
// clients can rely on receiving a JSON array.
type scanResponse struct {
	Sensitivity string   `json:"sensitivity"`
	Flags       []string `json:"flags"`
	Action      string   `json:"action"`
}

// ServeHTTP implements http.Handler.
func (h *Scan) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
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

	result, err := h.classifier.ClassifySensitivity(r.Context(), llm.Email{
		Subject:    req.Subject,
		Body:       req.Body,
		Recipients: req.To,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "sensitivity classification failed")
		return
	}

	flags := result.Reasons
	if flags == nil {
		flags = []string{}
	}
	writeJSON(w, http.StatusOK, scanResponse{
		Sensitivity: result.Level,
		Flags:       flags,
		Action:      actionForLevel(result.Level),
	})
}

// actionForLevel maps a sensitivity level to the action the proxy should take.
// Per the design, HIGH is redirected for review, MEDIUM is flagged, and LOW
// passes through untouched.
func actionForLevel(level string) string {
	switch level {
	case llm.SensitivityHigh:
		return ActionRedirect
	case llm.SensitivityMedium:
		return ActionFlag
	default:
		return ActionPass
	}
}
