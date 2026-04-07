package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_GeminiAPIKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "test-api-key")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.APIKey != "test-api-key" {
		t.Errorf("APIKey = %q, want %q", cfg.Provider.APIKey, "test-api-key")
	}
	if cfg.Provider.VertexProject != "" {
		t.Error("VertexProject should be empty for API key auth")
	}
	if cfg.Backend() != BackendGeminiAPI {
		t.Errorf("Backend() = %q, want %q", cfg.Backend(), BackendGeminiAPI)
	}
}

func TestLoad_GeminiAPIKeyFallback(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "fallback-key")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.APIKey != "fallback-key" {
		t.Errorf("APIKey = %q, want %q", cfg.Provider.APIKey, "fallback-key")
	}
}

func TestLoad_GoogleAPIKeyTakesPrecedence(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "primary-key")
	t.Setenv("GEMINI_API_KEY", "fallback-key")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.APIKey != "primary-key" {
		t.Errorf("APIKey = %q, want %q", cfg.Provider.APIKey, "primary-key")
	}
	if cfg.Backend() != BackendGeminiAPI {
		t.Errorf("Backend() = %q, want %q", cfg.Backend(), BackendGeminiAPI)
	}
}

func TestLoad_VertexAI(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "my-project")
	t.Setenv("GOOGLE_CLOUD_LOCATION", "europe-west1")
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.VertexProject != "my-project" {
		t.Errorf("VertexProject = %q, want %q", cfg.Provider.VertexProject, "my-project")
	}
	if cfg.Provider.VertexLocation != "europe-west1" {
		t.Errorf("VertexLocation = %q, want %q", cfg.Provider.VertexLocation, "europe-west1")
	}
	if cfg.Backend() != BackendVertexAI {
		t.Errorf("Backend() = %q, want %q", cfg.Backend(), BackendVertexAI)
	}
}

func TestLoad_VertexAIDefaultLocation(t *testing.T) {
	t.Setenv("GOOGLE_CLOUD_PROJECT", "my-project")
	t.Setenv("GOOGLE_CLOUD_LOCATION", "")
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.VertexLocation != "us-central1" {
		t.Errorf("VertexLocation = %q, want %q", cfg.Provider.VertexLocation, "us-central1")
	}
}

func TestLoad_NoCredentials(t *testing.T) {
	t.Setenv("MEDIA_OUTPUT_DIR", t.TempDir())
	// Use t.Setenv to blank out all credential vars — this ensures
	// automatic restoration after the test, preventing state leaks.
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "GEMINI_API_KEY") {
		t.Fatalf("error %q does not mention GEMINI_API_KEY", err)
	}
}

func TestConfigBackend_NilConfig(t *testing.T) {
	var cfg *Config
	if cfg.Backend() != BackendUnknown {
		t.Fatalf("Backend() = %q, want %q", cfg.Backend(), BackendUnknown)
	}
}

func TestLoad_DefaultOutputDir(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "test-key")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("MEDIA_OUTPUT_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, "generated_media")
	if cfg.OutputDir != expected {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, expected)
	}
}

func TestLoad_CustomOutputDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GOOGLE_API_KEY", "test-key")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	t.Setenv("MEDIA_OUTPUT_DIR", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != dir {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, dir)
	}
}
