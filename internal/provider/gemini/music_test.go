package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

func TestSaveMusic(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	data := []byte("fake-music-data")
	path, err := p.saveMusic(data, "mp3")
	if err != nil {
		t.Fatalf("saveMusic: %v", err)
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

func TestSaveMusic_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	path, err := p.saveMusic([]byte("data"), "mp3")
	if err != nil {
		t.Fatalf("saveMusic: %v", err)
	}

	base := filepath.Base(path)
	if ext := filepath.Ext(base); ext != ".mp3" {
		t.Errorf("extension = %q, want .mp3", ext)
	}
	if base[:6] != "music-" {
		t.Errorf("filename %q does not start with 'music-'", base)
	}
}

func TestGenerateMusic_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.GenerateMusic(context.Background(), provider.MusicRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestGenerateMusic_RejectsKnownNonMusicModel(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.GenerateMusic(context.Background(), provider.MusicRequest{
		Prompt: "a piano solo",
		Model:  "nb2",
	})
	if err == nil {
		t.Fatal("expected error for non-music model")
	}
	if err != nil && err.Error() != `model "gemini-3.1-flash-image-preview" does not support music generation` {
		t.Fatalf("unexpected error: %v", err)
	}
}
