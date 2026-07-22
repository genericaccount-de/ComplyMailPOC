package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/genericaccount-de/comply-mail-poc/backend/internal/llm"
)

// fakeClassifier is a test double for SensitivityClassifier.
type fakeClassifier struct {
	result llm.SensitivityResult
	err    error

	gotEmail llm.Email
	called   bool
}

func (f *fakeClassifier) ClassifySensitivity(_ context.Context, email llm.Email) (llm.SensitivityResult, error) {
	f.called = true
	f.gotEmail = email
	return f.result, f.err
}

func TestScan_HighRedirects(t *testing.T) {
	fake := &fakeClassifier{result: llm.SensitivityResult{
		Level:   llm.SensitivityHigh,
		Reasons: []string{"contains credentials"},
	}}
	h := NewScan(fake)

	rec := doRequest(t, h, `{"from":"a@x.com","to":["b@y.com"],"subject":"secret","body":"password is 123"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	var resp scanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Sensitivity != llm.SensitivityHigh {
		t.Errorf("sensitivity = %q, want HIGH", resp.Sensitivity)
	}
	if resp.Action != ActionRedirect {
		t.Errorf("action = %q, want %q", resp.Action, ActionRedirect)
	}
	if len(resp.Flags) != 1 || resp.Flags[0] != "contains credentials" {
		t.Errorf("unexpected flags: %+v", resp.Flags)
	}
	if !fake.called {
		t.Error("classifier was not called")
	}
	if fake.gotEmail.Body != "password is 123" {
		t.Errorf("unexpected email body passed: %q", fake.gotEmail.Body)
	}
}

func TestScan_ActionMapping(t *testing.T) {
	cases := map[string]string{
		llm.SensitivityLow:    ActionPass,
		llm.SensitivityMedium: ActionFlag,
		llm.SensitivityHigh:   ActionRedirect,
	}
	for level, wantAction := range cases {
		h := NewScan(&fakeClassifier{result: llm.SensitivityResult{Level: level}})
		rec := doRequest(t, h, `{"from":"a@x.com","to":["b@y.com"],"subject":"s","body":"b"}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("level %s: status = %d, want 200", level, rec.Code)
		}
		var resp scanResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("level %s: decode: %v", level, err)
		}
		if resp.Action != wantAction {
			t.Errorf("level %s: action = %q, want %q", level, resp.Action, wantAction)
		}
	}
}

func TestScan_EmptyFlagsIsArray(t *testing.T) {
	h := NewScan(&fakeClassifier{result: llm.SensitivityResult{Level: llm.SensitivityLow, Reasons: nil}})

	rec := doRequest(t, h, `{"from":"a@x.com","to":["b@y.com"],"subject":"s","body":"b"}`)

	if got := rec.Body.String(); !strings.Contains(got, `"flags":[]`) {
		t.Errorf("expected empty JSON array for flags, got %s", got)
	}
}

func TestScan_InvalidJSON(t *testing.T) {
	h := NewScan(&fakeClassifier{})

	rec := doRequest(t, h, `{not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestScan_MissingSubjectAndBody(t *testing.T) {
	fake := &fakeClassifier{}
	h := NewScan(fake)

	rec := doRequest(t, h, `{"from":"a@x.com","to":["b@y.com"],"subject":"  ","body":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if fake.called {
		t.Error("classifier should not be called when subject and body are empty")
	}
}

func TestScan_ClassifierError(t *testing.T) {
	h := NewScan(&fakeClassifier{err: errors.New("upstream down")})

	rec := doRequest(t, h, `{"from":"a@x.com","to":["b@y.com"],"subject":"s","body":"b"}`)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
