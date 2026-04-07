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

func TestValidateVideoGenerationInput_InvalidAspectRatio(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	err := p.validateVideoGenerationInput(p.modelMap["lite"], "1:1", "720p", 4)
	if err == nil || !strings.Contains(err.Error(), "aspectRatio") {
		t.Fatalf("expected aspectRatio validation error, got %v", err)
	}
}

func TestValidateVideoGenerationInput_InvalidResolution(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	err := p.validateVideoGenerationInput(p.modelMap["fast"], "16:9", "1440p", 4)
	if err == nil || !strings.Contains(err.Error(), "resolution") {
		t.Fatalf("expected resolution validation error, got %v", err)
	}
}

func TestValidateVideoGenerationInput_InvalidDuration(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	err := p.validateVideoGenerationInput(p.modelMap["fast"], "16:9", "1080p", 5)
	if err == nil || !strings.Contains(err.Error(), "duration") {
		t.Fatalf("expected duration validation error, got %v", err)
	}
}

func TestValidateVideoGenerationInput_LiteRejects4K(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	err := p.validateVideoGenerationInput(p.modelMap["lite"], "16:9", "4k", 4)
	if err == nil || !strings.Contains(err.Error(), "4k") {
		t.Fatalf("expected lite 4k validation error, got %v", err)
	}
}

func TestValidateVideoGenerationInput_RejectsNonVideoModel(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	err := p.validateVideoGenerationInput(p.modelMap["nb2"], "16:9", "720p", 4)
	if err == nil || !strings.Contains(err.Error(), "does not support video generation") {
		t.Fatalf("expected non-video model validation error, got %v", err)
	}
}

func TestValidateVideoGenerationInput_AllowsFast4K(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	if err := p.validateVideoGenerationInput(p.modelMap["fast"], "16:9", "4k", 8); err != nil {
		t.Fatalf("validateVideoGenerationInput: %v", err)
	}
}

func TestResolveExtensionModel_DefaultsToOriginalModel(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	got, err := p.resolveExtensionModel("models/veo-3.1-fast-generate-preview/operations/op-123", "")
	if err != nil {
		t.Fatalf("resolveExtensionModel: %v", err)
	}
	if got != p.modelMap["fast"] {
		t.Fatalf("model = %q, want %q", got, p.modelMap["fast"])
	}
}

func TestResolveExtensionModel_RejectsLite(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.resolveExtensionModel("models/veo-3.1-fast-generate-preview/operations/op-123", "lite")
	if err == nil || !strings.Contains(err.Error(), "does not support video extension") {
		t.Fatalf("expected lite extension validation error, got %v", err)
	}
}

func TestResolveExtensionModel_RejectsModelMismatch(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.resolveExtensionModel("models/veo-3.1-generate-preview/operations/op-123", "fast")
	if err == nil || !strings.Contains(err.Error(), "must match original model") {
		t.Fatalf("expected model mismatch error, got %v", err)
	}
}

func TestResolveExtensionModel_RequiresOriginalModel(t *testing.T) {
	p := &GeminiProvider{modelMap: defaultModelMap()}
	_, err := p.resolveExtensionModel("operations/op-123", "")
	if err == nil || !strings.Contains(err.Error(), "could not determine original model") {
		t.Fatalf("expected missing original model error, got %v", err)
	}
}

func TestModelFromOperationName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantModel string
	}{
		{
			name:      "parses model from operation name",
			input:     "models/veo-3.1-lite-generate-preview/operations/op-123",
			wantModel: "veo-3.1-lite-generate-preview",
		},
		{
			name:      "returns empty for malformed operation name",
			input:     "operations/op-123",
			wantModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modelFromOperationName(tt.input); got != tt.wantModel {
				t.Fatalf("modelFromOperationName(%q) = %q, want %q", tt.input, got, tt.wantModel)
			}
		})
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
