package main

import (
	"context"
	"log"

	"github.com/mordor-forge/gemini-media-mcp/internal/config"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider/gemini"
	"github.com/mordor-forge/gemini-media-mcp/internal/server"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	p, err := gemini.NewFromConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("creating provider: %v", err)
	}

	// GeminiProvider implements all three interfaces
	srv := server.New(p, p, p, cfg.OutputDir)

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
