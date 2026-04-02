//go:build e2e

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_OutputDirCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output", "dir")
	t.Setenv("GOOGLE_API_KEY", "test-key")
	t.Setenv("MEDIA_OUTPUT_DIR", dir)
	// Clear other credential vars
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.OutputDir != dir {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, dir)
	}
	// Verify the directory was actually created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("output dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("output path is not a directory")
	}
}
