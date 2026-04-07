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
	audio     provider.AudioGenerator
	music     provider.MusicGenerator
	models    provider.ModelLister
	backend   string
	outputDir string
}

// Options configures server metadata exposed via MCP tools.
type Options struct {
	Backend   string
	OutputDir string
}

// New creates a Server with the given provider implementations and output
// directory. Backend metadata defaults to "unknown". Any provider may be nil
// if that category of tools is not needed. Tools are only registered for
// non-nil providers.
func New(images provider.ImageGenerator, videos provider.VideoGenerator, audio provider.AudioGenerator, music provider.MusicGenerator, models provider.ModelLister, outputDir string) *Server {
	return NewWithOptions(images, videos, audio, music, models, Options{
		Backend:   "unknown",
		OutputDir: outputDir,
	})
}

// NewWithOptions creates a Server with explicit metadata about the configured
// backend and output directory.
func NewWithOptions(images provider.ImageGenerator, videos provider.VideoGenerator, audio provider.AudioGenerator, music provider.MusicGenerator, models provider.ModelLister, opts Options) *Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "gemini-media-mcp",
		Version: "0.1.0",
	}, nil)

	if opts.Backend == "" {
		opts.Backend = "unknown"
	}

	s := &Server{
		mcp:       mcpServer,
		images:    images,
		videos:    videos,
		audio:     audio,
		music:     music,
		models:    models,
		backend:   opts.Backend,
		outputDir: opts.OutputDir,
	}

	if images != nil {
		s.registerImageTools()
	}
	if videos != nil {
		s.registerVideoTools()
	}
	if audio != nil {
		s.registerAudioTools()
	}
	if music != nil {
		s.registerMusicTools()
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
