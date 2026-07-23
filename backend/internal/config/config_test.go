package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoad_ExpandsEnvVar(t *testing.T) {
	t.Setenv("LLM_API_KEY", "secret-123")
	path := writeConfig(t, `
server:
  listen_addr: ":9090"
llm:
  base_url: "https://openrouter.ai/api/v1"
  api_key: "${LLM_API_KEY}"
  model: "gpt-x"
  timeout: "15s"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.LLM.APIKey != "secret-123" {
		t.Errorf("api_key = %q, want expanded value", cfg.LLM.APIKey)
	}
	if cfg.Server.ListenAddr != ":9090" {
		t.Errorf("listen_addr = %q", cfg.Server.ListenAddr)
	}
	if time.Duration(cfg.LLM.Timeout) != 15*time.Second {
		t.Errorf("timeout = %v, want 15s", time.Duration(cfg.LLM.Timeout))
	}
}

func TestLoad_UnsetEnvVarFails(t *testing.T) {
	os.Unsetenv("MISSING_TOKEN")
	path := writeConfig(t, `
llm:
  base_url: "https://x/v1"
  api_key: "${MISSING_TOKEN}"
  model: "m"
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for unset env var, got nil")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	path := writeConfig(t, `
llm:
  base_url: "https://x/v1"
  model: "m"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.ListenAddr != DefaultListenAddr {
		t.Errorf("listen_addr = %q, want default %q", cfg.Server.ListenAddr, DefaultListenAddr)
	}
	if cfg.LLM.Timeout != DefaultLLMTimeout {
		t.Errorf("timeout = %v, want default", time.Duration(cfg.LLM.Timeout))
	}
}

func TestLoad_MissingRequiredFails(t *testing.T) {
	path := writeConfig(t, `
llm:
  api_key: "k"
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected validation error for missing base_url/model, got nil")
	}
}
