package gemini

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveModel(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}

	tests := []struct {
		name     string
		input    string
		fallback string
		want     string
	}{
		{"lite resolves", "lite", "standard", "veo-3.1-lite-generate-preview"},
		{"fast resolves", "fast", "lite", "veo-3.1-fast-generate-preview"},
		{"standard resolves", "standard", "lite", "veo-3.1-generate-preview"},
		{"nb2 resolves", "nb2", "pro", "gemini-3.1-flash-image-preview"},
		{"pro resolves", "pro", "nb2", "gemini-3-pro-image-preview"},
		{"empty uses fallback", "", "nb2", "gemini-3.1-flash-image-preview"},
		{"raw ID passes through", "veo-3.1-generate-preview", "lite", "veo-3.1-generate-preview"},
		{"unknown raw ID passes through", "custom-model-v1", "lite", "custom-model-v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.resolveModel(tt.input, tt.fallback)
			if got != tt.want {
				t.Errorf("resolveModel(%q, %q) = %q, want %q", tt.input, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		ext       string
		wantPfx   string
		wantExt   string
	}{
		{"image png", "image", "png", "image-", ".png"},
		{"video mp4", "video", "mp4", "video-", ".mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFilename(tt.mediaType, tt.ext)
			if !strings.HasPrefix(got, tt.wantPfx) {
				t.Errorf("generateFilename() = %q, want prefix %q", got, tt.wantPfx)
			}
			if filepath.Ext(got) != tt.wantExt {
				t.Errorf("generateFilename() ext = %q, want %q", filepath.Ext(got), tt.wantExt)
			}
		})
	}
}

func TestGenerateFilename_Uniqueness(t *testing.T) {
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateFilename("image", "png")
		if names[name] {
			t.Fatalf("duplicate filename generated: %s", name)
		}
		names[name] = true
	}
}
