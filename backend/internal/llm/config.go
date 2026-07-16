package llm

import "time"

// Default configuration values target a local Ollama instance, which
// exposes an OpenAI-compatible API. Point BaseURL at vLLM, llama-server,
// or Mistral La Plateforme (https://api.mistral.ai/v1) to switch providers.
const (
	DefaultBaseURL = "http://localhost:11434/v1"
	DefaultModel   = "mistral"
	DefaultTimeout = 30 * time.Second
)

// Environment variable names expected by the application's config loader
// (loading itself lives in internal/config, not here).
const (
	EnvBaseURL = "LLM_BASE_URL"
	EnvAPIKey  = "LLM_API_KEY"
	EnvModel   = "LLM_MODEL"
	EnvTimeout = "LLM_TIMEOUT"
)

// Config holds the settings needed to reach an OpenAI-compatible endpoint.
type Config struct {
	// BaseURL is the API root, e.g. "http://localhost:11434/v1".
	// The "/chat/completions" path is appended by the client.
	BaseURL string
	// APIKey is sent as a Bearer token. May be empty for local runtimes
	// such as Ollama that do not require authentication.
	APIKey string
	// Model is the model identifier, e.g. "mistral".
	Model string
	// Timeout bounds each HTTP request. Zero means DefaultTimeout.
	Timeout time.Duration
}

// withDefaults returns a copy of cfg with empty fields replaced by defaults.
func (c Config) withDefaults() Config {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.Model == "" {
		c.Model = DefaultModel
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	return c
}
