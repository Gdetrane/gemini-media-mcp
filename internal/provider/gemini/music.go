package gemini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/genai"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// GenerateMusic creates music from a text prompt using Gemini's
// GenerateContent API with audio response modality and Lyria models.
func (p *GeminiProvider) GenerateMusic(ctx context.Context, req provider.MusicRequest) (*provider.MusicResult, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	model := p.resolveModel(req.Model, "clip")
	if err := p.validateKnownModel(model, "music generation", p.modelMap["clip"], p.modelMap["full"]); err != nil {
		return nil, err
	}

	contents := []*genai.Content{
		genai.NewContentFromText(req.Prompt, genai.RoleUser),
	}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO", "TEXT"},
	}

	resp, err := p.client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("generating music: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no music returned by the API")
	}

	audioData, mimeType, lyrics := extractMusicResponse(resp)

	if len(audioData) == 0 {
		return nil, fmt.Errorf("no audio data found in API response")
	}

	ext := audioExtensionFromMIME(mimeType)
	filePath, err := p.saveMusic(audioData, ext)
	if err != nil {
		return nil, err
	}

	return &provider.MusicResult{
		FilePath: filePath,
		Model:    model,
		MimeType: mimeType,
		Lyrics:   lyrics,
	}, nil
}

// saveMusic writes raw audio bytes to the output directory with a generated filename.
func (p *GeminiProvider) saveMusic(data []byte, ext string) (string, error) {
	filename := generateFilename("music", ext)
	filePath := filepath.Join(p.outputDir, filename)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("writing file %s: %w", filePath, err)
	}
	return filePath, nil
}
