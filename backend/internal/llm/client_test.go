package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient spins up an httptest server whose handler is provided by the
// test, and returns a client pointed at it.
func newTestClient(t *testing.T, handler http.HandlerFunc) *HTTPClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL, Model: "test-model", Timeout: 2 * time.Second})
}

// chatReply writes a minimal OpenAI-compatible response whose single choice
// carries the given content.
func chatReply(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	resp := ChatResponse{
		ID:      "test",
		Model:   "test-model",
		Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: content}}},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		t.Fatalf("encode reply: %v", err)
	}
}

func TestChat_HappyPath(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %q", r.Method)
		}
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("expected model to default to test-model, got %q", req.Model)
		}
		chatReply(t, w, "hello")
	})

	resp, err := c.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if got := resp.Choices[0].Message.Content; got != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}
}

func TestChat_Non2xx(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	_, err := c.Chat(context.Background(), ChatRequest{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", apiErr.StatusCode)
	}
}

func TestChat_ContextCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		chatReply(t, w, "late")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.Chat(ctx, ChatRequest{})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

func TestClassifySensitivity_ParsesJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		chatReply(t, w, `{"level":"high","reasons":["contains credentials"]}`)
	})

	res, err := c.ClassifySensitivity(context.Background(), Email{Subject: "s", Body: "b"})
	if err != nil {
		t.Fatalf("ClassifySensitivity: %v", err)
	}
	if res.Level != SensitivityHigh {
		t.Errorf("level = %q, want HIGH (normalized)", res.Level)
	}
	if len(res.Reasons) != 1 {
		t.Errorf("reasons = %v, want 1 item", res.Reasons)
	}
}

func TestClassifySensitivity_InvalidLevel(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		chatReply(t, w, `{"level":"EXTREME","reasons":[]}`)
	})

	_, err := c.ClassifySensitivity(context.Background(), Email{})
	if err == nil {
		t.Fatal("expected error for invalid level, got nil")
	}
}

func TestClassifySensitivity_NonJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		chatReply(t, w, "I think this is medium risk.")
	})

	_, err := c.ClassifySensitivity(context.Background(), Email{})
	if err == nil {
		t.Fatal("expected parse error for non-JSON content, got nil")
	}
}

func TestAnalyzeStyle_ArrayShape(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		chatReply(t, w, `[{"type":"style","severity":"warning","message":"Greeting too informal"}]`)
	})

	got, err := c.AnalyzeStyle(context.Background(), "guide", Email{Subject: "s", Body: "hey"})
	if err != nil {
		t.Fatalf("AnalyzeStyle: %v", err)
	}
	if len(got) != 1 || got[0].Severity != "warning" {
		t.Errorf("unexpected suggestions: %+v", got)
	}
}

func TestAnalyzeStyle_ObjectShape(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		chatReply(t, w, `{"suggestions":[{"type":"style","severity":"info","message":"Add closing line"}]}`)
	})

	got, err := c.AnalyzeStyle(context.Background(), "", Email{})
	if err != nil {
		t.Fatalf("AnalyzeStyle: %v", err)
	}
	if len(got) != 1 || got[0].Message != "Add closing line" {
		t.Errorf("unexpected suggestions: %+v", got)
	}
}

func TestNew_Defaults(t *testing.T) {
	c := New(Config{})
	if c.baseURL != "http://localhost:11434/v1" {
		t.Errorf("baseURL = %q, want default Ollama URL", c.baseURL)
	}
	if c.Model() != DefaultModel {
		t.Errorf("model = %q, want %q", c.Model(), DefaultModel)
	}
}
