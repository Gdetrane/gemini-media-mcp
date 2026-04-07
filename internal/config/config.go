package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	BackendUnknown   = "unknown"
	BackendGeminiAPI = "gemini-api"
	BackendVertexAI  = "vertex-ai"
)

// Config holds all configuration for the MCP server.
type Config struct {
	Provider  ProviderConfig
	OutputDir string
}

// ProviderConfig holds authentication configuration.
type ProviderConfig struct {
	APIKey         string
	VertexProject  string
	VertexLocation string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		OutputDir: envOr("MEDIA_OUTPUT_DIR", defaultOutputDir()),
		Provider: ProviderConfig{
			APIKey:         firstEnv("GOOGLE_API_KEY", "GEMINI_API_KEY"),
			VertexProject:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
			VertexLocation: envOr("GOOGLE_CLOUD_LOCATION", "us-central1"),
		},
	}

	if cfg.Provider.APIKey == "" && cfg.Provider.VertexProject == "" {
		return nil, fmt.Errorf(
			"no credentials configured: set GOOGLE_API_KEY or GEMINI_API_KEY for Gemini API, " +
				"or GOOGLE_CLOUD_PROJECT for Vertex AI",
		)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output directory %s: %w", cfg.OutputDir, err)
	}

	return cfg, nil
}

// Backend returns the effective backend based on the configured credentials.
// API key auth takes precedence over Vertex AI to match provider construction.
func (c *Config) Backend() string {
	if c == nil {
		return BackendUnknown
	}
	if c.Provider.APIKey != "" {
		return BackendGeminiAPI
	}
	if c.Provider.VertexProject != "" {
		return BackendVertexAI
	}
	return BackendUnknown
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func defaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "generated_media"
	}
	return filepath.Join(home, "generated_media")
}
