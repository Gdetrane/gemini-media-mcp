//go:build e2e

package server

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/config"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider/gemini"
)

// connectE2EClient creates a Server backed by a real GeminiProvider,
// connects a test client via in-memory MCP transport, and returns the
// client session. This exercises the full stack: MCP protocol -> server
// -> provider -> genai SDK.
func connectE2EClient(t *testing.T) *mcp.ClientSession {
	t.Helper()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	outDir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	p, err := gemini.New(ctx, gemini.Config{
		APIKey:    apiKey,
		OutputDir: outDir,
	})
	if err != nil {
		t.Fatalf("creating gemini provider: %v", err)
	}

	srv := New(p, p, p, p, p, outDir)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "e2e-test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	return session
}

func TestE2E_MCPProtocol_AllToolsRegistered(t *testing.T) {
	session := connectE2EClient(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	// When all providers are wired (real GeminiProvider), all tools should
	// be registered.
	wantTools := map[string]bool{
		"generate_image": false,
		"edit_image":     false,
		"compose_images": false,
		"generate_video": false,
		"animate_image":  false,
		"extend_video":   false,
		"video_status":   false,
		"download_video": false,
		"generate_audio": false,
		"generate_music": false,
		"list_models":    false,
		"get_config":     false,
	}

	for _, tool := range result.Tools {
		t.Logf("Tool registered: %s", tool.Name)
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}

	for name, found := range wantTools {
		if !found {
			t.Errorf("tool %q not registered", name)
		}
	}

	t.Logf("Total tools registered: %d", len(result.Tools))
}

func TestE2E_MCPProtocol_ListModels(t *testing.T) {
	session := connectE2EClient(t)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_models",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool list_models: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error")
	}

	assertContentContains(t, res, "Available Models")
	assertContentContains(t, res, "gemini-3.1-flash-image-preview")
	assertContentContains(t, res, "veo-3.1-lite-generate-preview")
	assertContentContains(t, res, "lyria-3-clip-preview")
	assertContentContains(t, res, "gemini-2.5-flash-preview-tts")

	// Verify structured output has all 8 models
	if res.StructuredContent == nil {
		t.Fatal("structured content is nil")
	}

	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	models, ok := m["models"].([]any)
	if !ok {
		t.Fatal("structured content missing 'models' array")
	}
	if len(models) != 8 {
		t.Errorf("got %d models in structured output, want 8", len(models))
	}
}

func TestE2E_MCPProtocol_GetConfig(t *testing.T) {
	session := connectE2EClient(t)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_config",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool get_config: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error")
	}

	assertContentContains(t, res, "Backend: gemini-api")
	assertStructuredField(t, res, "backend", "gemini-api")
}

func TestE2E_ConfigLoad_MatchesProvider(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	// Verify config.Load produces values that NewFromConfig accepts.
	outDir := t.TempDir()
	t.Setenv("MEDIA_OUTPUT_DIR", outDir)
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	if cfg.Provider.APIKey != apiKey {
		t.Errorf("config APIKey = %q (len %d), want GOOGLE_API_KEY from env",
			cfg.Provider.APIKey[:4]+"...", len(cfg.Provider.APIKey))
	}

	ctx := context.Background()
	p, err := gemini.NewFromConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("NewFromConfig: %v", err)
	}

	// Verify provider works by calling ListModels
	models, err := p.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 8 {
		t.Errorf("got %d models, want 8", len(models))
	}
}
