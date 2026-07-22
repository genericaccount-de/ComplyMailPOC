package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/genericaccount-de/comply-mail-poc/backend/internal/llm"
)

// fakeAnalyzer is a test double for StyleAnalyzer.
type fakeAnalyzer struct {
	suggestions []llm.StyleSuggestion
	err         error

	gotStyleGuide string
	gotEmail      llm.Email
	called        bool
}

func (f *fakeAnalyzer) AnalyzeStyle(_ context.Context, styleGuide string, email llm.Email) ([]llm.StyleSuggestion, error) {
	f.called = true
	f.gotStyleGuide = styleGuide
	f.gotEmail = email
	return f.suggestions, f.err
}

func doRequest(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/check-style-email", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCompose_HappyPath(t *testing.T) {
	fake := &fakeAnalyzer{suggestions: []llm.StyleSuggestion{
		{Type: "style", Severity: "warning", Message: "Greeting too informal"},
	}}
	h := NewStyleCheck(fake, "")

	rec := doRequest(t, h, `{"subject":"hi","body":"hey there","recipients":["a@x.com"]}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var resp styleCheckResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Suggestions) != 1 || resp.Suggestions[0].Message != "Greeting too informal" {
		t.Errorf("unexpected suggestions: %+v", resp.Suggestions)
	}
	if !fake.called {
		t.Error("analyzer was not called")
	}
	if fake.gotStyleGuide == "" {
		t.Error("expected default style guide to be passed")
	}
	if fake.gotEmail.Subject != "hi" || fake.gotEmail.Body != "hey there" {
		t.Errorf("unexpected email passed to analyzer: %+v", fake.gotEmail)
	}
}

func TestCompose_EmptySuggestionsIsArray(t *testing.T) {
	h := NewStyleCheck(&fakeAnalyzer{suggestions: nil}, "")

	rec := doRequest(t, h, `{"subject":"s","body":"b"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); !strings.Contains(got, `"suggestions":[]`) {
		t.Errorf("expected empty JSON array, got %s", got)
	}
}

func TestCompose_InvalidJSON(t *testing.T) {
	h := NewStyleCheck(&fakeAnalyzer{}, "")

	rec := doRequest(t, h, `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCompose_MissingSubjectAndBody(t *testing.T) {
	fake := &fakeAnalyzer{}
	h := NewStyleCheck(fake, "")

	rec := doRequest(t, h, `{"subject":"   ","body":"","recipients":["a@x.com"]}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if fake.called {
		t.Error("analyzer should not be called when subject and body are empty")
	}
}

func TestCompose_AnalyzerError(t *testing.T) {
	h := NewStyleCheck(&fakeAnalyzer{err: errors.New("upstream down")}, "")

	rec := doRequest(t, h, `{"subject":"s","body":"b"}`)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestCompose_CustomStyleGuide(t *testing.T) {
	fake := &fakeAnalyzer{}
	h := NewStyleCheck(fake, "MY GUIDE")

	doRequest(t, h, `{"subject":"s","body":"b"}`)

	if fake.gotStyleGuide != "MY GUIDE" {
		t.Errorf("style guide = %q, want %q", fake.gotStyleGuide, "MY GUIDE")
	}
}
