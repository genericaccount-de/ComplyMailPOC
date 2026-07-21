// Package config loads and validates the SMTP proxy configuration
// from a YAML file (listen addr, upstream host/port, backend API URL).
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level SMTP proxy configuration.
type Config struct {
	// ListenAddr is the address the proxy listens on, e.g. ":2525".
	ListenAddr string         `yaml:"listen_addr"`
	Upstream   UpstreamConfig `yaml:"upstream"`
	Backend    BackendConfig  `yaml:"backend"`
}

// UpstreamConfig points at the customer's real mail server.
type UpstreamConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// BackendConfig points at the ComplyMail backend API.
type BackendConfig struct {
	// APIURL is the base URL used to reach the backend, e.g. "http://api:8080".
	APIURL string `yaml:"api_url"`
}

// Default values applied when the corresponding field is empty.
const (
	DefaultListenAddr   = ":2525"
	DefaultUpstreamPort = 25
)

// Load reads, parses, and validates the YAML config at path, applying defaults.
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
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.ListenAddr == "" {
		c.ListenAddr = DefaultListenAddr
	}
	if c.Upstream.Port == 0 {
		c.Upstream.Port = DefaultUpstreamPort
	}
}

func (c *Config) validate() error {
	if c.Upstream.Host == "" {
		return fmt.Errorf("config: upstream.host is required")
	}
	if c.Backend.APIURL == "" {
		return fmt.Errorf("config: backend.api_url is required")
	}
	return nil
}
