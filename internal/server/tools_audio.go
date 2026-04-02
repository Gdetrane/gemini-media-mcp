package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// registerAudioTools adds the generate_audio tool to the MCP server.
func (s *Server) registerAudioTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "generate_audio",
		Description: "Generate speech audio from a text prompt using Google's Gemini TTS. Supports voice selection and language configuration.",
	}, s.handleGenerateAudio)
}

func (s *Server) handleGenerateAudio(ctx context.Context, _ *mcp.CallToolRequest, input provider.AudioRequest) (*mcp.CallToolResult, provider.AudioResult, error) {
	result, err := s.audio.GenerateAudio(ctx, input)
	if err != nil {
		return nil, provider.AudioResult{}, fmt.Errorf("generate audio: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Audio generated!\n\nModel: %s\nSaved to: %s\nType: %s",
				result.Model, result.FilePath, result.MimeType,
			)},
		},
	}, *result, nil
}
