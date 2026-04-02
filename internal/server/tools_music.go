package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// registerMusicTools adds the generate_music tool to the MCP server.
func (s *Server) registerMusicTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "generate_music",
		Description: "Generate music from a text prompt using Google's Lyria models. Supports genre, instruments, BPM, key, mood, structure tags like [Verse] [Chorus] [Bridge], and custom lyrics.",
	}, s.handleGenerateMusic)
}

func (s *Server) handleGenerateMusic(ctx context.Context, _ *mcp.CallToolRequest, input provider.MusicRequest) (*mcp.CallToolResult, provider.MusicResult, error) {
	result, err := s.music.GenerateMusic(ctx, input)
	if err != nil {
		return nil, provider.MusicResult{}, fmt.Errorf("generate music: %w", err)
	}

	text := fmt.Sprintf(
		"Music generated!\n\nModel: %s\nSaved to: %s\nType: %s",
		result.Model, result.FilePath, result.MimeType,
	)
	if result.Lyrics != "" {
		text += fmt.Sprintf("\n\nLyrics/Structure:\n%s", result.Lyrics)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, *result, nil
}
