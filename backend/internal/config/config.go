// Package config loads and validates the backend API configuration
// from a YAML file. A single secret (the LLM API key) may be overridden
// via the LLM_API_KEY environment variable so credentials can be kept out
// of the config file when desired.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// EnvAPIKey is the environment variable that overrides llm.api_key.
const EnvAPIKey = "LLM_API_KEY"

// Config is the top-level backend configuration.
type Config struct {
	Server ServerConfig `yaml:"server"`
	LLM    LLMConfig    `yaml:"llm"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	// ListenAddr is the address the API listens on, e.g. ":8080".
	ListenAddr string `yaml:"listen_addr"`
}

// LLMConfig holds settings for the OpenAI-compatible LLM endpoint.
type LLMConfig struct {
	// BaseURL is the API root, e.g. "https://api.mistral.ai/v1".
	BaseURL string `yaml:"base_url"`
	// APIKey is sent as a Bearer token. May be empty for local runtimes.
	APIKey string `yaml:"api_key"`
	// Model is the model identifier, e.g. "mistral-small-latest".
	Model string `yaml:"model"`
	// Timeout bounds each HTTP request, e.g. "30s".
	Timeout Duration `yaml:"timeout"`
}

// Duration is a time.Duration that unmarshals from a human-readable YAML
// string such as "30s" or "1m30s".
type Duration time.Duration

// UnmarshalYAML parses a duration string (see time.ParseDuration).
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("config: invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// Default values applied when the corresponding field is empty.
const (
	DefaultListenAddr = ":8080"
	DefaultLLMTimeout = Duration(30 * time.Second)
)

// Load reads, parses, and validates the YAML config at path, applying
// defaults and environment overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	cfg.applyDefaults()
	cfg.applyEnvOverrides()

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.ListenAddr == "" {
		c.Server.ListenAddr = DefaultListenAddr
	}
	if c.LLM.Timeout == 0 {
		c.LLM.Timeout = DefaultLLMTimeout
	}
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv(EnvAPIKey); v != "" {
		c.LLM.APIKey = v
	}
}

func (c *Config) validate() error {
	if c.LLM.BaseURL == "" {
		return fmt.Errorf("config: llm.base_url is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("config: llm.model is required")
	}
	return nil
}
