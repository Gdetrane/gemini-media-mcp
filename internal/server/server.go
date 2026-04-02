// Package server implements the MCP server that exposes Gemini media
// generation capabilities as MCP tools over stdio transport.
package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// Server wraps an MCP server and routes tool calls to provider implementations.
type Server struct {
	mcp       *mcp.Server
	images    provider.ImageGenerator
	videos    provider.VideoGenerator
	models    provider.ModelLister
	outputDir string
}

// New creates a Server with the given provider implementations and output
// directory. Any provider may be nil if that category of tools is not needed.
// Tools are only registered for non-nil providers.
func New(images provider.ImageGenerator, videos provider.VideoGenerator, models provider.ModelLister, outputDir string) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "gemini-media-mcp",
		Version: "0.1.0",
	}, nil)

	s := &Server{
		mcp:       mcpServer,
		images:    images,
		videos:    videos,
		models:    models,
		outputDir: outputDir,
	}

	if images != nil {
		s.registerImageTools()
	}
	if videos != nil {
		s.registerVideoTools()
	}
	s.registerConfigTools()

	return s
}

// Run starts the MCP server on the stdio transport, blocking until the
// client disconnects or the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// MCPServer returns the underlying mcp.Server for testing or advanced usage.
func (s *Server) MCPServer() *mcp.Server {
	return s.mcp
}
