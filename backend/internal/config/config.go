// Package config loads and validates the backend API configuration
// from a YAML file.
//
// Any value may reference an environment variable using ${VAR} or $VAR
// syntax (see os.Expand); references are resolved when the file is loaded and
// loading fails if a referenced variable is unset. This keeps secrets such as
// the LLM API key out of the file, e.g. api_key: "${LLM_API_KEY}".
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

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

// Load reads, expands ${VAR} references, parses, and validates the YAML
// config at path, applying defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	expanded, err := expandEnv(string(data))
	if err != nil {
		return nil, fmt.Errorf("config: %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// expandEnv replaces ${VAR}/$VAR references in s with the corresponding
// environment variable values. It returns an error listing any referenced
// variables that are not set, so misconfiguration fails fast at startup.
func expandEnv(s string) (string, error) {
	var missing []string
	expanded := os.Expand(s, func(name string) string {
		v, ok := os.LookupEnv(name)
		if !ok {
			missing = append(missing, name)
			return ""
		}
		return v
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("unset environment variable(s): %s", strings.Join(missing, ", "))
	}
	return expanded, nil
}

func (c *Config) applyDefaults() {
	if c.Server.ListenAddr == "" {
		c.Server.ListenAddr = DefaultListenAddr
	}
	if c.LLM.Timeout == 0 {
		c.LLM.Timeout = DefaultLLMTimeout
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
