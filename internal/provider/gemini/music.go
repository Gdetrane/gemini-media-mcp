package gemini

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no music returned by the API")
	}

	var audioData []byte
	var mimeType string
	var lyrics strings.Builder

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != nil {
			audioData = part.InlineData.Data
			mimeType = part.InlineData.MIMEType
		}
		if part.Text != "" {
			if lyrics.Len() > 0 {
				lyrics.WriteString("\n")
			}
			lyrics.WriteString(part.Text)
		}
	}

	if audioData == nil {
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
		Lyrics:   lyrics.String(),
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
