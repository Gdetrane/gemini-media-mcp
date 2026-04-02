//go:build e2e

package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// newTestProvider creates a GeminiProvider configured for E2E tests.
// It skips the test if GOOGLE_API_KEY is not set and uses t.TempDir()
// as the output directory so generated files are cleaned up automatically.
func newTestProvider(t *testing.T) *GeminiProvider {
	t.Helper()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	p, err := New(context.Background(), Config{
		APIKey:    apiKey,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating provider: %v", err)
	}
	return p
}

// assertImageResult validates the common invariants of a successful image
// generation: non-empty path, file exists, file is non-zero size.
// It returns the os.FileInfo for further inspection.
func assertImageResult(t *testing.T, result *provider.ImageResult) os.FileInfo {
	t.Helper()

	if result.FilePath == "" {
		t.Fatal("result has empty FilePath")
	}
	if result.Model == "" {
		t.Error("result has empty Model")
	}
	if result.MimeType == "" {
		t.Error("result has empty MimeType")
	}

	fi, err := os.Stat(result.FilePath)
	if err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatal("generated file has zero size")
	}

	t.Logf("path=%s model=%s mime=%s size=%d bytes", result.FilePath, result.Model, result.MimeType, fi.Size())
	return fi
}

func TestE2E_ImageGenerate(t *testing.T) {
	tests := []struct {
		name    string
		req     provider.ImageRequest
		wantErr bool
		// checkModel is the expected resolved model ID substring (empty = skip check).
		checkModel string
	}{
		{
			name: "default_parameters",
			req: provider.ImageRequest{
				Prompt: "a solid red square on white background",
			},
			checkModel: "gemini-3.1-flash-image-preview",
		},
		{
			name: "explicit_nb2_model",
			req: provider.ImageRequest{
				Prompt: "a solid blue circle on white background",
				Model:  "nb2",
			},
			checkModel: "gemini-3.1-flash-image-preview",
		},
		{
			name: "pro_model",
			req: provider.ImageRequest{
				Prompt: "a solid green triangle on white background",
				Model:  "pro",
			},
			checkModel: "gemini-3-pro-image-preview",
		},
		{
			name: "aspect_ratio_16_9",
			req: provider.ImageRequest{
				Prompt:      "a solid yellow rectangle on white background",
				AspectRatio: "16:9",
			},
		},
		{
			name: "aspect_ratio_9_16",
			req: provider.ImageRequest{
				Prompt:      "a solid orange rectangle on white background",
				AspectRatio: "9:16",
			},
		},
		{
			name: "aspect_ratio_1_1",
			req: provider.ImageRequest{
				Prompt:      "a solid purple square on white background",
				AspectRatio: "1:1",
			},
		},
		{
			name:    "empty_prompt_error",
			req:     provider.ImageRequest{Prompt: ""},
			wantErr: true,
		},
		{
			name: "raw_model_id",
			req: provider.ImageRequest{
				Prompt: "a solid cyan diamond on white background",
				Model:  "gemini-3.1-flash-image-preview",
			},
			checkModel: "gemini-3.1-flash-image-preview",
		},
	}

	p := newTestProvider(t)
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := p.Generate(ctx, tc.req)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				t.Logf("got expected error: %v", err)
				return
			}

			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			fi := assertImageResult(t, result)

			if tc.checkModel != "" && result.Model != tc.checkModel {
				t.Errorf("model mismatch: got %q, want %q", result.Model, tc.checkModel)
			}

			t.Logf("file size: %d bytes", fi.Size())
		})
	}
}

func TestE2E_ImageEdit(t *testing.T) {
	p := newTestProvider(t)
	ctx := context.Background()

	// Step 1: Generate a source image to edit.
	t.Log("generating source image for edit test...")
	source, err := p.Generate(ctx, provider.ImageRequest{
		Prompt: "a solid red square on white background",
	})
	if err != nil {
		t.Fatalf("generating source image: %v", err)
	}
	assertImageResult(t, source)
	t.Logf("source image ready: %s", source.FilePath)

	// Step 2: Edit the generated image.
	t.Log("editing source image...")
	edited, err := p.Edit(ctx, provider.EditImageRequest{
		Prompt:    "change the red square to blue and add a yellow border",
		ImagePath: source.FilePath,
	})
	if err != nil {
		t.Fatalf("editing image: %v", err)
	}

	fi := assertImageResult(t, edited)

	if edited.FilePath == source.FilePath {
		t.Error("edited image has the same path as source -- expected a new file")
	}

	t.Logf("edited image: %s (%d bytes)", edited.FilePath, fi.Size())
}
