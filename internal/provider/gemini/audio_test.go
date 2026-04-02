package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

func TestSaveAudio(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	data := []byte("fake-audio-data")
	path, err := p.saveAudio(data, "wav")
	if err != nil {
		t.Fatalf("saveAudio: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("saved to %q, want dir %q", path, dir)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("file content = %q, want %q", content, data)
	}
}

func TestSaveAudio_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	path, err := p.saveAudio([]byte("data"), "mp3")
	if err != nil {
		t.Fatalf("saveAudio: %v", err)
	}

	base := filepath.Base(path)
	if ext := filepath.Ext(base); ext != ".mp3" {
		t.Errorf("extension = %q, want .mp3", ext)
	}
	if base[:6] != "audio-" {
		t.Errorf("filename %q does not start with 'audio-'", base)
	}
}

func TestGenerateAudio_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.GenerateAudio(context.Background(), provider.AudioRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestAudioExtensionFromMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"audio/wav", "wav"},
		{"audio/x-wav", "wav"},
		{"audio/mpeg", "mp3"},
		{"audio/mp3", "mp3"},
		{"audio/ogg", "ogg"},
		{"audio/flac", "flac"},
		{"audio/aac", "aac"},
		{"audio/unknown", "wav"},
		{"", "wav"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			got := audioExtensionFromMIME(tt.mime)
			if got != tt.want {
				t.Errorf("audioExtensionFromMIME(%q) = %q, want %q", tt.mime, got, tt.want)
			}
		})
	}
}
