//go:build e2e

package gemini

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

func TestE2E_GenerateImage(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	p, err := New(ctx, Config{
		APIKey:    apiKey,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating provider: %v", err)
	}

	result, err := p.Generate(ctx, provider.ImageRequest{
		Prompt: "a simple red circle on a white background",
		Model:  "nb2",
	})
	if err != nil {
		t.Fatalf("generating image: %v", err)
	}

	if result.FilePath == "" {
		t.Error("empty file path")
	}
	if _, err := os.Stat(result.FilePath); err != nil {
		t.Errorf("file not found: %v", err)
	}
	t.Logf("Generated image: %s (model: %s)", result.FilePath, result.Model)
}

func TestE2E_GenerateVideo(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	p, err := New(ctx, Config{
		APIKey:    apiKey,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating provider: %v", err)
	}

	// Start generation (cheapest: lite, 4s, 720p)
	op, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt:     "a simple animation of a bouncing red ball on white background",
		Model:      "lite",
		Duration:   4,
		Resolution: "720p",
	})
	if err != nil {
		t.Fatalf("starting video generation: %v", err)
	}
	t.Logf("Video operation started: %s", op.OperationID)

	// Poll until done (timeout after 5 minutes)
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		status, err := p.Status(ctx, op.OperationID)
		if err != nil {
			t.Fatalf("checking status: %v", err)
		}
		t.Logf("Status: %s (done: %v)", status.Progress, status.Done)

		if status.Done {
			if status.Progress == "failed" {
				t.Fatalf("video generation failed: %s", status.Error)
			}
			break
		}
		time.Sleep(10 * time.Second)
	}

	// Download
	result, err := p.Download(ctx, op.OperationID)
	if err != nil {
		t.Fatalf("downloading video: %v", err)
	}

	if result.FilePath == "" {
		t.Error("empty file path")
	}
	if _, err := os.Stat(result.FilePath); err != nil {
		t.Errorf("file not found: %v", err)
	}
	t.Logf("Downloaded video: %s", result.FilePath)
}
