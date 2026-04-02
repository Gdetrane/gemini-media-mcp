//go:build e2e

package gemini

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

const (
	videoPollInterval = 10 * time.Second
	videoPollTimeout  = 5 * time.Minute
	videoMinSize      = 10 * 1024 // 10 KB — any valid video will exceed this
	videoPrompt       = "a bouncing red ball on white background"
)

// newVideoTestProvider creates a GeminiProvider configured for video E2E tests.
// It skips the test if GOOGLE_API_KEY is not set.
func newVideoTestProvider(t *testing.T) *GeminiProvider {
	t.Helper()
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
	return p
}

// pollUntilDone polls an operation every videoPollInterval until it finishes
// or the timeout expires. It logs each status check and fails the test on
// errors or timeout.
func pollUntilDone(t *testing.T, p *GeminiProvider, operationID string) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(videoPollTimeout)
	for time.Now().Before(deadline) {
		status, err := p.Status(ctx, operationID)
		if err != nil {
			t.Fatalf("polling status for %s: %v", operationID, err)
		}
		t.Logf("  status: %s (done=%v)", status.Progress, status.Done)
		if status.Done {
			if status.Progress == "failed" {
				t.Fatalf("video generation failed: %s", status.Error)
			}
			return
		}
		time.Sleep(videoPollInterval)
	}
	t.Fatalf("timed out after %v waiting for operation %s", videoPollTimeout, operationID)
}

// verifyVideoFile asserts that the file at path exists, is non-empty, and
// exceeds videoMinSize. It logs the path and size.
func verifyVideoFile(t *testing.T, path, label string) {
	t.Helper()
	if path == "" {
		t.Fatal("file path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatal("video file has zero size")
	}
	if info.Size() < int64(videoMinSize) {
		t.Fatalf("video file too small (%d bytes), expected at least %d", info.Size(), videoMinSize)
	}
	t.Logf("[%s] path=%s size=%d bytes", label, path, info.Size())
}

// TestE2E_Video_Landscape tests the full round trip: GenerateVideo with lite
// model, 4s duration, 720p, 16:9 aspect ratio. It polls until complete, then
// downloads and verifies the MP4 file.
func TestE2E_Video_Landscape(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	op, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt:      videoPrompt,
		Model:       "lite",
		Duration:    4,
		Resolution:  "720p",
		AspectRatio: "16:9",
	})
	if err != nil {
		t.Fatalf("GenerateVideo: %v", err)
	}
	t.Logf("operation started: %s (model: %s)", op.OperationID, op.Model)

	pollUntilDone(t, p, op.OperationID)

	result, err := p.Download(ctx, op.OperationID)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	verifyVideoFile(t, result.FilePath, "landscape")
	t.Logf("model=%s duration=%d operationID=%s", result.Model, result.Duration, result.OperationID)
}

// TestE2E_Video_Portrait tests portrait orientation (9:16) with lite/4s/720p.
func TestE2E_Video_Portrait(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	op, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt:      videoPrompt,
		Model:       "lite",
		Duration:    4,
		Resolution:  "720p",
		AspectRatio: "9:16",
	})
	if err != nil {
		t.Fatalf("GenerateVideo: %v", err)
	}
	t.Logf("operation started: %s (model: %s)", op.OperationID, op.Model)

	pollUntilDone(t, p, op.OperationID)

	result, err := p.Download(ctx, op.OperationID)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	verifyVideoFile(t, result.FilePath, "portrait")
	t.Logf("model=%s duration=%d operationID=%s", result.Model, result.Duration, result.OperationID)
}

// TestE2E_Video_StatusPolling verifies that Status returns meaningful progress
// strings. It starts a generation and checks that at least one poll returns
// "processing" or "complete".
func TestE2E_Video_StatusPolling(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	op, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt:     videoPrompt,
		Model:      "lite",
		Duration:   4,
		Resolution: "720p",
	})
	if err != nil {
		t.Fatalf("GenerateVideo: %v", err)
	}
	t.Logf("operation started: %s", op.OperationID)

	validProgress := map[string]bool{
		"pending":    true,
		"processing": true,
		"complete":   true,
		"failed":     true,
	}

	sawMeaningful := false
	deadline := time.Now().Add(videoPollTimeout)
	for time.Now().Before(deadline) {
		status, err := p.Status(ctx, op.OperationID)
		if err != nil {
			t.Fatalf("Status: %v", err)
		}
		t.Logf("  progress=%q done=%v", status.Progress, status.Done)

		if !validProgress[status.Progress] {
			t.Errorf("unexpected progress value: %q", status.Progress)
		}
		if status.Progress == "processing" || status.Progress == "complete" {
			sawMeaningful = true
		}
		if status.Done {
			break
		}
		time.Sleep(videoPollInterval)
	}

	if !sawMeaningful {
		t.Error("never saw 'processing' or 'complete' status during polling")
	}
}

// TestE2E_Video_DownloadBeforeComplete starts a generation and immediately
// tries to download without waiting. It verifies that Download returns an
// error indicating the operation is not yet complete.
func TestE2E_Video_DownloadBeforeComplete(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	op, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt:     videoPrompt,
		Model:      "lite",
		Duration:   4,
		Resolution: "720p",
	})
	if err != nil {
		t.Fatalf("GenerateVideo: %v", err)
	}
	t.Logf("operation started: %s, attempting immediate download", op.OperationID)

	_, err = p.Download(ctx, op.OperationID)
	if err == nil {
		t.Fatal("expected error downloading before complete, got nil")
	}
	if !strings.Contains(err.Error(), "not yet complete") {
		t.Fatalf("expected 'not yet complete' error, got: %v", err)
	}
	t.Logf("got expected error: %v", err)
}

// TestE2E_Video_EmptyPrompt verifies that GenerateVideo returns a local
// validation error when the prompt is empty, without making an API call.
func TestE2E_Video_EmptyPrompt(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	_, err := p.GenerateVideo(ctx, provider.VideoRequest{
		Prompt: "",
		Model:  "lite",
	})
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
	if !strings.Contains(err.Error(), "prompt is required") {
		t.Fatalf("expected 'prompt is required' error, got: %v", err)
	}
	t.Logf("got expected error: %v", err)
}

// TestE2E_Video_AnimateImage tests the image-to-video pipeline. It first
// generates an image using the image provider, then passes it to AnimateImage
// to create a video from it. Uses lite model, 4s duration.
func TestE2E_Video_AnimateImage(t *testing.T) {
	p := newVideoTestProvider(t)
	ctx := context.Background()

	// Step 1: Generate a source image.
	t.Log("generating source image...")
	imgResult, err := p.Generate(ctx, provider.ImageRequest{
		Prompt: "a simple red circle on a white background",
		Model:  "nb2",
	})
	if err != nil {
		t.Fatalf("Generate image: %v", err)
	}
	t.Logf("source image: %s (model: %s)", imgResult.FilePath, imgResult.Model)

	// Step 2: Animate the image into a video.
	t.Log("animating image into video...")
	op, err := p.AnimateImage(ctx, provider.AnimateRequest{
		Prompt:    "the red circle gently bouncing up and down",
		ImagePath: imgResult.FilePath,
		Model:     "lite",
		Duration:  4,
	})
	if err != nil {
		t.Fatalf("AnimateImage: %v", err)
	}
	t.Logf("operation started: %s (model: %s)", op.OperationID, op.Model)

	// Step 3: Poll until done.
	pollUntilDone(t, p, op.OperationID)

	// Step 4: Download and verify.
	result, err := p.Download(ctx, op.OperationID)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	verifyVideoFile(t, result.FilePath, "animate-image")
	t.Logf("model=%s duration=%d operationID=%s", result.Model, result.Duration, result.OperationID)
}
