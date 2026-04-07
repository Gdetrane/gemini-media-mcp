package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// --- Mock providers ---

type mockImageGen struct {
	generateResult *provider.ImageResult
	generateErr    error
	editResult     *provider.ImageResult
	editErr        error
	composeResult  *provider.ImageResult
	composeErr     error
}

func (m *mockImageGen) Generate(_ context.Context, _ provider.ImageRequest) (*provider.ImageResult, error) {
	return m.generateResult, m.generateErr
}

func (m *mockImageGen) Edit(_ context.Context, _ provider.EditImageRequest) (*provider.ImageResult, error) {
	return m.editResult, m.editErr
}

func (m *mockImageGen) Compose(_ context.Context, _ provider.ComposeRequest) (*provider.ImageResult, error) {
	return m.composeResult, m.composeErr
}

type mockVideoGen struct {
	generateResult *provider.VideoOperation
	generateErr    error
	animateResult  *provider.VideoOperation
	animateErr     error
	extendResult   *provider.VideoOperation
	extendErr      error
	statusResult   *provider.VideoStatus
	statusErr      error
	downloadResult *provider.VideoResult
	downloadErr    error
}

func (m *mockVideoGen) GenerateVideo(_ context.Context, _ provider.VideoRequest) (*provider.VideoOperation, error) {
	return m.generateResult, m.generateErr
}

func (m *mockVideoGen) AnimateImage(_ context.Context, _ provider.AnimateRequest) (*provider.VideoOperation, error) {
	return m.animateResult, m.animateErr
}

func (m *mockVideoGen) Extend(_ context.Context, _ provider.ExtendRequest) (*provider.VideoOperation, error) {
	return m.extendResult, m.extendErr
}

func (m *mockVideoGen) Status(_ context.Context, _ string) (*provider.VideoStatus, error) {
	return m.statusResult, m.statusErr
}

func (m *mockVideoGen) Download(_ context.Context, _ string) (*provider.VideoResult, error) {
	return m.downloadResult, m.downloadErr
}

type mockAudioGen struct {
	generateResult *provider.AudioResult
	generateErr    error
}

func (m *mockAudioGen) GenerateAudio(_ context.Context, _ provider.AudioRequest) (*provider.AudioResult, error) {
	return m.generateResult, m.generateErr
}

type mockMusicGen struct {
	generateResult *provider.MusicResult
	generateErr    error
}

func (m *mockMusicGen) GenerateMusic(_ context.Context, _ provider.MusicRequest) (*provider.MusicResult, error) {
	return m.generateResult, m.generateErr
}

type mockModelLister struct {
	models []provider.ModelInfo
	err    error
}

func (m *mockModelLister) ListModels(_ context.Context) ([]provider.ModelInfo, error) {
	return m.models, m.err
}

// --- Constructor tests ---

func TestNew_CreatesServer(t *testing.T) {
	srv := New(&mockImageGen{}, nil, nil, nil, nil, t.TempDir())
	if srv == nil {
		t.Fatal("New returned nil")
	}
	if srv.mcp == nil {
		t.Fatal("underlying MCP server is nil")
	}
}

func TestNew_NilProviders(t *testing.T) {
	srv := New(nil, nil, nil, nil, nil, t.TempDir())
	if srv == nil {
		t.Fatal("New returned nil with all nil providers")
	}
}

func TestMCPServer_ReturnsUnderlyingServer(t *testing.T) {
	srv := New(&mockImageGen{}, nil, nil, nil, nil, t.TempDir())
	if srv.MCPServer() != srv.mcp {
		t.Fatal("MCPServer() did not return the underlying server")
	}
}

// --- Tool handler tests via in-memory MCP transport ---

// connectTestClient creates a Server with the given mock, connects a test
// client via in-memory transport, and returns the client session. The server
// runs in a background goroutine tied to the test context.
func connectTestClient(t *testing.T, mock *mockImageGen) *mcp.ClientSession {
	t.Helper()

	srv := New(mock, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
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

func TestToolsRegistered(t *testing.T) {
	session := connectTestClient(t, &mockImageGen{
		generateResult: &provider.ImageResult{},
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	// Image tools + get_config (always registered). Video/model tools are
	// not expected because connectTestClient passes nil for those providers.
	wantTools := map[string]bool{
		"generate_image": false,
		"edit_image":     false,
		"compose_images": false,
		"get_config":     false,
	}
	for _, tool := range result.Tools {
		if _, ok := wantTools[tool.Name]; ok {
			wantTools[tool.Name] = true
		}
	}
	for name, found := range wantTools {
		if !found {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestToolsNotRegistered_WhenImageProviderNil(t *testing.T) {
	srv := New(nil, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	// get_config is always registered, but image/video/list_models tools
	// should be absent when their providers are nil.
	if len(result.Tools) != 1 {
		names := make([]string, len(result.Tools))
		for i, tool := range result.Tools {
			names[i] = tool.Name
		}
		t.Errorf("expected 1 tool (get_config) when all providers nil, got %d: %v", len(result.Tools), names)
	}
	if len(result.Tools) == 1 && result.Tools[0].Name != "get_config" {
		t.Errorf("expected get_config, got %q", result.Tools[0].Name)
	}
}

func TestGenerateImage_Success(t *testing.T) {
	mock := &mockImageGen{
		generateResult: &provider.ImageResult{
			FilePath: "/tmp/test/image-abc.png",
			Model:    "gemini-3.1-flash-image-preview",
			MimeType: "image/png",
		},
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_image",
		Arguments: map[string]any{
			"prompt": "a sunset over mountains",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	// Check text content
	assertContentContains(t, res, "Image generated!")
	assertContentContains(t, res, "gemini-3.1-flash-image-preview")
	assertContentContains(t, res, "/tmp/test/image-abc.png")

	// Check structured output
	assertStructuredField(t, res, "filePath", "/tmp/test/image-abc.png")
}

func TestGenerateImage_Error(t *testing.T) {
	mock := &mockImageGen{
		generateErr: errors.New("API rate limit exceeded"),
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_image",
		Arguments: map[string]any{
			"prompt": "anything",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "API rate limit exceeded")
}

func TestGenerateImage_MissingPrompt(t *testing.T) {
	session := connectTestClient(t, &mockImageGen{
		generateResult: &provider.ImageResult{},
	})

	// The SDK validates required fields at the protocol level and returns
	// a Go error from CallTool (not an IsError result).
	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "generate_image",
		Arguments: map[string]any{},
	})
	if err == nil {
		t.Fatal("expected validation error for missing required prompt, got nil")
	}
	if !contains(err.Error(), "required") && !contains(err.Error(), "prompt") {
		t.Errorf("error message %q does not mention missing prompt", err.Error())
	}
}

func TestEditImage_Success(t *testing.T) {
	mock := &mockImageGen{
		editResult: &provider.ImageResult{
			FilePath: "/tmp/test/edited-xyz.png",
			Model:    "gemini-3-pro-image-preview",
			MimeType: "image/png",
		},
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "edit_image",
		Arguments: map[string]any{
			"prompt":    "make the sky purple",
			"imagePath": "/tmp/source.png",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Image edited!")
	assertContentContains(t, res, "gemini-3-pro-image-preview")
	assertStructuredField(t, res, "filePath", "/tmp/test/edited-xyz.png")
}

func TestEditImage_Error(t *testing.T) {
	mock := &mockImageGen{
		editErr: errors.New("source image not found"),
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "edit_image",
		Arguments: map[string]any{
			"prompt":    "change colors",
			"imagePath": "/nonexistent.png",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "source image not found")
}

func TestComposeImages_Success(t *testing.T) {
	mock := &mockImageGen{
		composeResult: &provider.ImageResult{
			FilePath: "/tmp/test/composed-123.png",
			Model:    "gemini-3.1-flash-image-preview",
			MimeType: "image/png",
		},
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "compose_images",
		Arguments: map[string]any{
			"prompt":          "combine these into a collage",
			"referenceImages": []string{"/tmp/a.png", "/tmp/b.png"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Image composed!")
	assertStructuredField(t, res, "filePath", "/tmp/test/composed-123.png")
}

func TestComposeImages_Error(t *testing.T) {
	mock := &mockImageGen{
		composeErr: errors.New("too many reference images"),
	}

	session := connectTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "compose_images",
		Arguments: map[string]any{
			"prompt":          "merge styles",
			"referenceImages": []string{"/tmp/a.png"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "too many reference images")
}

// --- Video tool tests ---

// connectVideoTestClient creates a Server with the given video mock,
// connects a test client via in-memory transport, and returns the session.
func connectVideoTestClient(t *testing.T, mock *mockVideoGen) *mcp.ClientSession {
	t.Helper()

	srv := New(nil, mock, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
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

func TestGenerateVideo_Success(t *testing.T) {
	mock := &mockVideoGen{
		generateResult: &provider.VideoOperation{
			OperationID: "op-video-123",
			Model:       "veo-2.0-generate-001",
		},
	}

	session := connectVideoTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_video",
		Arguments: map[string]any{
			"prompt": "a cat chasing a laser pointer",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Video generation started!")
	assertContentContains(t, res, "op-video-123")
	assertContentContains(t, res, "veo-2.0-generate-001")
	assertStructuredField(t, res, "operationId", "op-video-123")
}

func TestVideoStatus_Success(t *testing.T) {
	mock := &mockVideoGen{
		statusResult: &provider.VideoStatus{
			OperationID: "op-video-123",
			Done:        false,
			Progress:    "processing",
		},
	}

	session := connectVideoTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "video_status",
		Arguments: map[string]any{
			"operationId": "op-video-123",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "op-video-123")
	assertContentContains(t, res, "processing")
	assertStructuredField(t, res, "operationId", "op-video-123")
	assertStructuredField(t, res, "progress", "processing")
}

func TestDownloadVideo_Success(t *testing.T) {
	mock := &mockVideoGen{
		downloadResult: &provider.VideoResult{
			FilePath:    "/tmp/test/video-abc.mp4",
			OperationID: "op-video-123",
			Model:       "veo-2.0-generate-001",
			Duration:    8,
		},
	}

	session := connectVideoTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "download_video",
		Arguments: map[string]any{
			"operationId": "op-video-123",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Video downloaded!")
	assertContentContains(t, res, "/tmp/test/video-abc.mp4")
	assertContentContains(t, res, "veo-2.0-generate-001")
	assertStructuredField(t, res, "filePath", "/tmp/test/video-abc.mp4")
	assertStructuredField(t, res, "operationId", "op-video-123")
}

func TestDownloadVideo_OmitsUnknownMetadata(t *testing.T) {
	mock := &mockVideoGen{
		downloadResult: &provider.VideoResult{
			FilePath:    "/tmp/test/video-abc.mp4",
			OperationID: "op-video-123",
		},
	}

	session := connectVideoTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "download_video",
		Arguments: map[string]any{
			"operationId": "op-video-123",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Video downloaded!")
	assertContentContains(t, res, "Operation: op-video-123")
	assertContentNotContains(t, res, "Model:")
	assertContentNotContains(t, res, "Duration:")
}

func TestVideoToolsNotRegistered_WhenVideoProviderNil(t *testing.T) {
	// Only image provider, no video provider.
	srv := New(&mockImageGen{}, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	videoTools := []string{"generate_video", "animate_image", "extend_video", "video_status", "download_video"}
	registered := make(map[string]bool)
	for _, tool := range result.Tools {
		registered[tool.Name] = true
	}
	for _, name := range videoTools {
		if registered[name] {
			t.Errorf("video tool %q should not be registered when video provider is nil", name)
		}
	}
}

// --- Config tool tests ---

func TestListModels_Success(t *testing.T) {
	lister := &mockModelLister{
		models: []provider.ModelInfo{
			{
				ID:           "gemini-3.1-flash-image-preview",
				Tier:         "nb2",
				MediaType:    "image",
				Resolutions:  []string{"1K", "2K"},
				AspectRatios: []string{"1:1", "16:9"},
				PricePerSec:  "$0.067/img",
			},
			{
				ID:          "veo-2.0-generate-001",
				Tier:        "standard",
				MediaType:   "video",
				PricePerSec: "$0.35",
			},
		},
	}

	srv := NewWithOptions(nil, nil, nil, nil, lister, Options{
		Backend:   "gemini-api",
		OutputDir: t.TempDir(),
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_models",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Available Models")
	assertContentContains(t, res, "gemini-3.1-flash-image-preview")
	assertContentContains(t, res, "veo-2.0-generate-001")
	assertContentContains(t, res, "$0.067/img")
	assertContentContains(t, res, "$0.35")
	assertContentNotContains(t, res, "/s")
}

func TestGetConfig_ReturnsBackendInfo(t *testing.T) {
	outDir := t.TempDir()
	srv := NewWithOptions(nil, nil, nil, nil, nil, Options{
		Backend:   "gemini-api",
		OutputDir: outDir,
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_config",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Backend: gemini-api")
	assertContentContains(t, res, outDir)
	assertStructuredField(t, res, "backend", "gemini-api")
	assertStructuredField(t, res, "outputDir", outDir)
}

func TestGetConfig_ReturnsVertexBackendInfo(t *testing.T) {
	outDir := t.TempDir()
	srv := NewWithOptions(nil, nil, nil, nil, nil, Options{
		Backend:   "vertex-ai",
		OutputDir: outDir,
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_config",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Backend: vertex-ai")
	assertStructuredField(t, res, "backend", "vertex-ai")
}

func TestListModelsNotRegistered_WhenModelListerNil(t *testing.T) {
	srv := New(nil, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "list_models" {
			t.Error("list_models should not be registered when model lister is nil")
		}
	}
}

// --- Audio tool tests ---

// connectAudioTestClient creates a Server with the given audio mock,
// connects a test client via in-memory transport, and returns the session.
func connectAudioTestClient(t *testing.T, mock *mockAudioGen) *mcp.ClientSession {
	t.Helper()

	srv := New(nil, nil, mock, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
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

func TestGenerateAudio_Success(t *testing.T) {
	mock := &mockAudioGen{
		generateResult: &provider.AudioResult{
			FilePath: "/tmp/test/audio-abc.wav",
			Model:    "gemini-2.5-flash",
			MimeType: "audio/wav",
		},
	}

	session := connectAudioTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_audio",
		Arguments: map[string]any{
			"prompt": "Say hello world in a cheerful voice",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Audio generated!")
	assertContentContains(t, res, "gemini-2.5-flash")
	assertContentContains(t, res, "/tmp/test/audio-abc.wav")
	assertStructuredField(t, res, "filePath", "/tmp/test/audio-abc.wav")
}

func TestGenerateAudio_Error(t *testing.T) {
	mock := &mockAudioGen{
		generateErr: errors.New("TTS quota exceeded"),
	}

	session := connectAudioTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_audio",
		Arguments: map[string]any{
			"prompt": "anything",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "TTS quota exceeded")
}

func TestAudioToolsNotRegistered_WhenAudioProviderNil(t *testing.T) {
	// Only image provider, no audio provider.
	srv := New(&mockImageGen{}, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "generate_audio" {
			t.Error("generate_audio should not be registered when audio provider is nil")
		}
	}
}

// --- Music tool tests ---

// connectMusicTestClient creates a Server with the given music mock,
// connects a test client via in-memory transport, and returns the session.
func connectMusicTestClient(t *testing.T, mock *mockMusicGen) *mcp.ClientSession {
	t.Helper()

	srv := New(nil, nil, nil, mock, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
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

func TestGenerateMusic_Success(t *testing.T) {
	mock := &mockMusicGen{
		generateResult: &provider.MusicResult{
			FilePath: "/tmp/test/music-abc.mp3",
			Model:    "lyria-3-clip-preview",
			MimeType: "audio/mp3",
		},
	}

	session := connectMusicTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_music",
		Arguments: map[string]any{
			"prompt": "A gentle acoustic guitar melody in C major",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Music generated!")
	assertContentContains(t, res, "lyria-3-clip-preview")
	assertContentContains(t, res, "/tmp/test/music-abc.mp3")
	assertStructuredField(t, res, "filePath", "/tmp/test/music-abc.mp3")
}

func TestGenerateMusic_WithLyrics(t *testing.T) {
	mock := &mockMusicGen{
		generateResult: &provider.MusicResult{
			FilePath: "/tmp/test/music-xyz.mp3",
			Model:    "lyria-3-pro-preview",
			MimeType: "audio/mp3",
			Lyrics:   "[Verse]\nHello world\n[Chorus]\nLa la la",
		},
	}

	session := connectMusicTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_music",
		Arguments: map[string]any{
			"prompt": "A pop song about coding",
			"model":  "full",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if res.IsError {
		t.Fatal("expected success, got error result")
	}

	assertContentContains(t, res, "Music generated!")
	assertContentContains(t, res, "Lyrics/Structure")
	assertStructuredField(t, res, "lyrics", "[Verse]\nHello world\n[Chorus]\nLa la la")
}

func TestGenerateMusic_Error(t *testing.T) {
	mock := &mockMusicGen{
		generateErr: errors.New("music generation quota exceeded"),
	}

	session := connectMusicTestClient(t, mock)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "generate_music",
		Arguments: map[string]any{
			"prompt": "anything",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	if !res.IsError {
		t.Fatal("expected error result, got success")
	}
	assertContentContains(t, res, "music generation quota exceeded")
}

func TestMusicToolsNotRegistered_WhenMusicProviderNil(t *testing.T) {
	// Only image provider, no music provider.
	srv := New(&mockImageGen{}, nil, nil, nil, nil, t.TempDir())

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.mcp.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
	})

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.Name == "generate_music" {
			t.Error("generate_music should not be registered when music provider is nil")
		}
	}
}

// --- Test helpers ---

// assertContentContains checks that at least one content entry in the result
// contains the given substring. It marshals each Content item to find text.
func assertContentContains(t *testing.T, res *mcp.CallToolResult, substr string) {
	t.Helper()
	for _, c := range res.Content {
		data, err := json.Marshal(c)
		if err != nil {
			continue
		}
		if contains(string(data), substr) {
			return
		}
	}
	t.Errorf("no content entry contains %q", substr)
}

func assertContentNotContains(t *testing.T, res *mcp.CallToolResult, substr string) {
	t.Helper()
	for _, c := range res.Content {
		data, err := json.Marshal(c)
		if err != nil {
			continue
		}
		if contains(string(data), substr) {
			t.Errorf("content entry unexpectedly contains %q", substr)
		}
	}
}

// assertStructuredField checks that the structured output contains a field
// with the expected value.
func assertStructuredField(t *testing.T, res *mcp.CallToolResult, key, want string) {
	t.Helper()
	if res.StructuredContent == nil {
		t.Fatalf("structured content is nil, expected field %q=%q", key, want)
	}

	data, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}

	got, ok := m[key]
	if !ok {
		t.Errorf("structured content missing field %q", key)
		return
	}
	if fmt, ok := got.(string); ok && fmt != want {
		t.Errorf("structured content[%q] = %q, want %q", key, fmt, want)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
