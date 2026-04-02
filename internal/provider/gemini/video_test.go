package gemini

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

func TestSaveVideo(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	data := []byte("fake-mp4-data")
	path, err := p.saveVideo(data)
	if err != nil {
		t.Fatalf("saveVideo: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("saved to %q, want dir %q", path, dir)
	}
	if ext := filepath.Ext(path); ext != ".mp4" {
		t.Errorf("extension = %q, want .mp4", ext)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("file content = %q, want %q", content, data)
	}
}

func TestSaveVideo_FilenameFormat(t *testing.T) {
	dir := t.TempDir()
	p := &GeminiProvider{outputDir: dir, modelMap: defaultModelMap()}

	path, err := p.saveVideo([]byte("data"))
	if err != nil {
		t.Fatalf("saveVideo: %v", err)
	}

	base := filepath.Base(path)
	if !strings.HasPrefix(base, "video-") {
		t.Errorf("filename %q does not start with 'video-'", base)
	}
	if ext := filepath.Ext(base); ext != ".mp4" {
		t.Errorf("extension = %q, want .mp4", ext)
	}
}

func TestGenerateVideo_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.GenerateVideo(context.Background(), provider.VideoRequest{Prompt: ""})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestAnimateImage_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.AnimateImage(context.Background(), provider.AnimateRequest{
		Prompt:    "",
		ImagePath: "/tmp/img.png",
	})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestAnimateImage_EmptyImagePath(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.AnimateImage(context.Background(), provider.AnimateRequest{
		Prompt:    "animate",
		ImagePath: "",
	})
	if err == nil {
		t.Fatal("expected error for empty imagePath")
	}
}

func TestAnimateImage_MissingImage(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.AnimateImage(context.Background(), provider.AnimateRequest{
		Prompt:    "animate",
		ImagePath: "/nonexistent/image.png",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent image")
	}
	if !strings.Contains(err.Error(), "reading source image") {
		t.Errorf("error = %q, want to contain 'reading source image'", err)
	}
}

func TestExtend_EmptyPrompt(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Extend(context.Background(), provider.ExtendRequest{
		Prompt:      "",
		OperationID: "op-123",
	})
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestExtend_EmptyOperationID(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Extend(context.Background(), provider.ExtendRequest{
		Prompt:      "continue",
		OperationID: "",
	})
	if err == nil {
		t.Fatal("expected error for empty operationId")
	}
}

func TestStatus_EmptyOperationID(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Status(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty operationId")
	}
}

func TestDownload_EmptyOperationID(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.Download(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty operationId")
	}
}

func TestBuildVideoConfig_Defaults(t *testing.T) {
	config := buildVideoConfig("", "", 0)
	if config.AspectRatio != "" {
		t.Errorf("AspectRatio = %q, want empty", config.AspectRatio)
	}
	if config.Resolution != "" {
		t.Errorf("Resolution = %q, want empty", config.Resolution)
	}
	if config.DurationSeconds != nil {
		t.Errorf("DurationSeconds = %v, want nil", config.DurationSeconds)
	}
}

func TestBuildVideoConfig_AllFields(t *testing.T) {
	config := buildVideoConfig("16:9", "1080p", 8)
	if config.AspectRatio != "16:9" {
		t.Errorf("AspectRatio = %q, want '16:9'", config.AspectRatio)
	}
	if config.Resolution != "1080p" {
		t.Errorf("Resolution = %q, want '1080p'", config.Resolution)
	}
	if config.DurationSeconds == nil {
		t.Fatal("DurationSeconds is nil, want 8")
	}
	if *config.DurationSeconds != 8 {
		t.Errorf("DurationSeconds = %d, want 8", *config.DurationSeconds)
	}
}

func TestBuildVideoConfig_DurationOnly(t *testing.T) {
	config := buildVideoConfig("", "", 4)
	if config.DurationSeconds == nil {
		t.Fatal("DurationSeconds is nil, want 4")
	}
	if *config.DurationSeconds != 4 {
		t.Errorf("DurationSeconds = %d, want 4", *config.DurationSeconds)
	}
}
