package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// EmptyInput is used for tools that require no parameters.
type EmptyInput struct{}

// registerConfigTools adds list_models and get_config tools to the MCP
// server. list_models requires a non-nil ModelLister; get_config is always
// available since it uses the server's own fields.
func (s *Server) registerConfigTools() {
	if s.models != nil {
		mcp.AddTool(s.mcp, &mcp.Tool{
			Name:        "list_models",
			Description: "List the models supported by this server with their tiers, capabilities, supported resolutions, and pricing guidance.",
		}, s.handleListModels)
	}

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_config",
		Description: "Show current server configuration including active backend and output directory.",
	}, s.handleGetConfig)
}

// modelsResult wraps the model list for structured output, since the MCP
// SDK requires output schemas to have type "object".
type modelsResult struct {
	Models []provider.ModelInfo `json:"models"`
}

// configResult is the structured output for get_config.
type configResult struct {
	Backend   string `json:"backend"`
	OutputDir string `json:"outputDir"`
}

func (s *Server) handleListModels(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, modelsResult, error) {
	models, err := s.models.ListModels(ctx)
	if err != nil {
		return nil, modelsResult{}, fmt.Errorf("list models: %w", err)
	}

	var b strings.Builder
	b.WriteString("Available Models\n")
	b.WriteString("================\n\n")
	for _, m := range models {
		fmt.Fprintf(&b, "%-40s  tier=%-8s  type=%-6s", m.ID, m.Tier, m.MediaType)
		if m.PricePerSec != "" {
			fmt.Fprintf(&b, "  price=%s", m.PricePerSec)
		}
		if len(m.Resolutions) > 0 {
			fmt.Fprintf(&b, "  res=[%s]", strings.Join(m.Resolutions, ", "))
		}
		if len(m.AspectRatios) > 0 {
			fmt.Fprintf(&b, "  ratios=[%s]", strings.Join(m.AspectRatios, ", "))
		}
		b.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: b.String()},
		},
	}, modelsResult{Models: models}, nil
}

func (s *Server) handleGetConfig(_ context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, configResult, error) {
	result := configResult{
		Backend:   s.backend,
		OutputDir: s.outputDir,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Backend: %s\nOutput directory: %s",
				result.Backend, result.OutputDir,
			)},
		},
	}, result, nil
}
