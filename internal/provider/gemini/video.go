package gemini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/genai"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// GenerateVideo creates a video from a text prompt using the Gemini Veo API.
// Video generation is asynchronous — this returns an operation handle that
// must be polled via Status and retrieved via Download.
func (p *GeminiProvider) GenerateVideo(ctx context.Context, req provider.VideoRequest) (*provider.VideoOperation, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	model := p.resolveModel(req.Model, "lite")

	config := buildVideoConfig(req.AspectRatio, req.Resolution, req.Duration)

	op, err := p.client.Models.GenerateVideosFromSource(ctx, model, &genai.GenerateVideosSource{
		Prompt: req.Prompt,
	}, config)
	if err != nil {
		return nil, fmt.Errorf("generating video: %w", err)
	}

	return &provider.VideoOperation{
		OperationID: op.Name,
		Model:       model,
	}, nil
}

// AnimateImage creates a video using an image as the first frame.
// The image is read from disk and sent alongside the text prompt.
func (p *GeminiProvider) AnimateImage(ctx context.Context, req provider.AnimateRequest) (*provider.VideoOperation, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if req.ImagePath == "" {
		return nil, fmt.Errorf("imagePath is required")
	}

	imgBytes, err := os.ReadFile(req.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("reading source image %s: %w", req.ImagePath, err)
	}

	model := p.resolveModel(req.Model, "lite")

	config := buildVideoConfig(req.AspectRatio, "", req.Duration)

	op, err := p.client.Models.GenerateVideosFromSource(ctx, model, &genai.GenerateVideosSource{
		Prompt: req.Prompt,
		Image: &genai.Image{
			ImageBytes: imgBytes,
			MIMEType:   mimeFromPath(req.ImagePath),
		},
	}, config)
	if err != nil {
		return nil, fmt.Errorf("animating image: %w", err)
	}

	return &provider.VideoOperation{
		OperationID: op.Name,
		Model:       model,
	}, nil
}

// Extend chains a new video segment onto a previously generated video.
// It retrieves the completed operation to get its video, then uses it as
// the source for a new generation. Lite model does not support extension.
func (p *GeminiProvider) Extend(ctx context.Context, req provider.ExtendRequest) (*provider.VideoOperation, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if req.OperationID == "" {
		return nil, fmt.Errorf("operationId is required")
	}

	// Retrieve the previous operation to get its video output.
	prevOp, err := p.client.Operations.GetVideosOperation(ctx, &genai.GenerateVideosOperation{
		Name: req.OperationID,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("retrieving previous operation %s: %w", req.OperationID, err)
	}
	if !prevOp.Done {
		return nil, fmt.Errorf("previous operation %s is not yet complete", req.OperationID)
	}
	if prevOp.Error != nil {
		return nil, fmt.Errorf("previous operation %s failed: %v", req.OperationID, prevOp.Error)
	}
	if prevOp.Response == nil || len(prevOp.Response.GeneratedVideos) == 0 || prevOp.Response.GeneratedVideos[0].Video == nil {
		return nil, fmt.Errorf("previous operation %s has no video output", req.OperationID)
	}

	prevVideo := prevOp.Response.GeneratedVideos[0].Video

	// Default to "fast" for extension — Lite does not support it.
	model := p.resolveModel(req.Model, "fast")

	op, err := p.client.Models.GenerateVideosFromSource(ctx, model, &genai.GenerateVideosSource{
		Prompt: req.Prompt,
		Video:  prevVideo,
	}, &genai.GenerateVideosConfig{})
	if err != nil {
		return nil, fmt.Errorf("extending video: %w", err)
	}

	return &provider.VideoOperation{
		OperationID: op.Name,
		Model:       model,
	}, nil
}

// Status polls the current state of a video generation operation.
func (p *GeminiProvider) Status(ctx context.Context, operationID string) (*provider.VideoStatus, error) {
	if operationID == "" {
		return nil, fmt.Errorf("operationId is required")
	}

	op, err := p.client.Operations.GetVideosOperation(ctx, &genai.GenerateVideosOperation{
		Name: operationID,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("polling operation %s: %w", operationID, err)
	}

	status := &provider.VideoStatus{
		OperationID: operationID,
		Done:        op.Done,
	}

	switch {
	case op.Error != nil:
		status.Progress = "failed"
		status.Error = fmt.Sprintf("%v", op.Error)
	case op.Done:
		status.Progress = "complete"
	default:
		status.Progress = "processing"
	}

	return status, nil
}

// Download retrieves a completed video operation and saves the video to disk.
func (p *GeminiProvider) Download(ctx context.Context, operationID string) (*provider.VideoResult, error) {
	if operationID == "" {
		return nil, fmt.Errorf("operationId is required")
	}

	op, err := p.client.Operations.GetVideosOperation(ctx, &genai.GenerateVideosOperation{
		Name: operationID,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("retrieving operation %s: %w", operationID, err)
	}

	if !op.Done {
		return nil, fmt.Errorf("operation %s is not yet complete", operationID)
	}
	if op.Error != nil {
		return nil, fmt.Errorf("operation %s failed: %v", operationID, op.Error)
	}
	if op.Response == nil || len(op.Response.GeneratedVideos) == 0 || op.Response.GeneratedVideos[0].Video == nil {
		return nil, fmt.Errorf("operation %s has no video output", operationID)
	}

	video := op.Response.GeneratedVideos[0].Video

	// The API returns a URI, not inline bytes — download via the SDK
	videoBytes := video.VideoBytes
	if len(videoBytes) == 0 && video.URI != "" {
		downloaded, err := p.client.Files.Download(ctx, genai.NewDownloadURIFromVideo(video), nil)
		if err != nil {
			return nil, fmt.Errorf("downloading video from URI: %w", err)
		}
		videoBytes = downloaded
	}
	if len(videoBytes) == 0 {
		return nil, fmt.Errorf("operation %s returned empty video data", operationID)
	}

	filePath, err := p.saveVideo(videoBytes)
	if err != nil {
		return nil, err
	}

	return &provider.VideoResult{
		FilePath:    filePath,
		OperationID: operationID,
		Model:       "", // model is not returned by the operation response
	}, nil
}

// saveVideo writes raw video bytes to the output directory with a generated filename.
func (p *GeminiProvider) saveVideo(data []byte) (string, error) {
	filename := generateFilename("video", "mp4")
	filePath := filepath.Join(p.outputDir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing file %s: %w", filePath, err)
	}
	return filePath, nil
}

// buildVideoConfig creates a GenerateVideosConfig from the optional parameters.
func buildVideoConfig(aspectRatio, resolution string, duration int) *genai.GenerateVideosConfig {
	config := &genai.GenerateVideosConfig{}
	if aspectRatio != "" {
		config.AspectRatio = aspectRatio
	}
	if resolution != "" {
		config.Resolution = resolution
	}
	if duration > 0 {
		d := int32(duration)
		config.DurationSeconds = &d
	}
	return config
}
