// Package config loads and validates the Mortar server configuration.
// Secrets are never stored in the config file directly; use ${VAR} syntax
// to reference environment variables for credentials and API keys.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top-level Mortar server configuration.
type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Plugins []PluginConfig `yaml:"plugins"`
	Routing RoutingConfig  `yaml:"routing"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port           int      `yaml:"port"`
	AllowedOrigins []string `yaml:"allowed_origins,omitempty"`
}

// PluginConfig holds the configuration for a single plugin instance.
// Secrets (api_key, password, etc.) should use ${VAR} interpolation.
type PluginConfig struct {
	ID       string `yaml:"id"`
	Type     string `yaml:"type"`
	URL      string `yaml:"url"`
	APIKey   string `yaml:"api_key,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// RoutingConfig holds request routing policy as resolved at startup.
// Keys in Requests correspond to capability names (video, audio, ebook).
type RoutingConfig struct {
	Requests RequestsRoutingConfig `yaml:"requests"`
}

// RequestsRoutingConfig maps each request capability to the plugin id
// that should handle it. Omit a key when zero or one compatible plugin
// exists — Mortar resolves that case automatically.
type RequestsRoutingConfig struct {
	Video string `yaml:"video,omitempty"`
	Audio string `yaml:"audio,omitempty"`
	Ebook string `yaml:"ebook,omitempty"`
}

var envVarRE = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads a YAML config file, interpolates ${VAR} expressions with
// environment variable values, and returns the parsed Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	// Pre-process: replace ${VAR} in the raw YAML text before parsing so
	// that the substituted values are parsed as their natural YAML types.
	// Collect names of any referenced variables that are unset.
	var missingVars []string
	raw := envVarRE.ReplaceAllStringFunc(string(data), func(match string) string {
		name := match[2 : len(match)-1]
		val := os.Getenv(name)
		if val == "" {
			missingVars = append(missingVars, name)
		}
		return val
	})
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("config: the following environment variables are referenced but not set: %s", strings.Join(missingVars, ", "))
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 3000
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate checks for obviously invalid configuration.
func validate(cfg *Config) error {
	seen := make(map[string]bool, len(cfg.Plugins))
	for _, p := range cfg.Plugins {
		if strings.TrimSpace(p.ID) == "" {
			return fmt.Errorf("config: plugin entry is missing an id")
		}
		if strings.TrimSpace(p.Type) == "" {
			return fmt.Errorf("config: plugin %q is missing a type", p.ID)
		}
		if strings.TrimSpace(p.URL) == "" {
			return fmt.Errorf("config: plugin %q is missing a url", p.ID)
		}
		if seen[p.ID] {
			return fmt.Errorf("config: duplicate plugin id %q", p.ID)
		}
		seen[p.ID] = true
	}

	pluginIDs := make(map[string]bool, len(cfg.Plugins))
	for _, p := range cfg.Plugins {
		pluginIDs[p.ID] = true
	}
	for cap, id := range map[string]string{
		"routing.requests.video": cfg.Routing.Requests.Video,
		"routing.requests.audio": cfg.Routing.Requests.Audio,
		"routing.requests.ebook": cfg.Routing.Requests.Ebook,
	} {
		if id != "" && !pluginIDs[id] {
			return fmt.Errorf("config: %s references unknown plugin id %q", cap, id)
		}
	}

	return nil
}
