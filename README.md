# gemini-media-mcp

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Unified Go MCP server for AI media generation via Google Gemini API and Vertex AI.

## Features

- **Image generation** -- text-to-image with configurable aspect ratios and resolutions (1K/2K/4K)
- **Image editing** -- modify existing images with natural language prompts
- **Multi-reference composition** -- combine up to 3 reference images with style/content guidance
- **Video generation** -- text-to-video via Veo 3.1 Lite, Fast, and Standard tiers
- **Image-to-video** -- animate still images into video clips
- **Video extension** -- chain clips for longer content (Fast and Standard tiers)
- **Single binary** -- no runtime dependencies, runs over stdio transport
- **Provider abstraction** -- backend-agnostic interfaces for image, video, and model operations
- **Dual backend** -- supports both Gemini API (API key) and Vertex AI (project credentials)

## Quick Start

```bash
# Install
go install github.com/mordor-forge/gemini-media-mcp/cmd/gemini-media-mcp@latest

# Configure (Gemini API)
export GOOGLE_API_KEY="your-api-key"

# Or configure (Vertex AI)
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_LOCATION="us-central1"

# Run directly (stdio transport)
gemini-media-mcp
```

Then add it to your MCP client -- see [MCP Client Configuration](#mcp-client-configuration) below.

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_API_KEY` | Yes* | -- | Gemini API key. `GEMINI_API_KEY` is also accepted |
| `GOOGLE_CLOUD_PROJECT` | Yes* | -- | GCP project ID for Vertex AI backend |
| `GOOGLE_CLOUD_LOCATION` | No | `us-central1` | GCP region for Vertex AI |
| `MEDIA_OUTPUT_DIR` | No | `~/generated_media` | Directory for saved media files |

*One of `GOOGLE_API_KEY` or `GOOGLE_CLOUD_PROJECT` must be set. If both are set, Vertex AI takes precedence.

## Available Tools

| Tool | Description | Type |
|------|-------------|------|
| `generate_image` | Generate image from text prompt | Sync |
| `edit_image` | Edit existing image with text prompt | Sync |
| `compose_images` | Multi-reference image composition (up to 3) | Sync |
| `generate_video` | Generate video from text prompt (returns operation ID) | Async |
| `animate_image` | Animate image into video (first frame) | Async |
| `extend_video` | Chain video clips for longer content | Async |
| `video_status` | Check video generation progress | Sync |
| `download_video` | Download completed video | Sync |
| `list_models` | Show available models with capabilities and pricing | Sync |
| `get_config` | Show current backend and configuration | Sync |

Async tools return an operation ID immediately. Use `video_status` to poll for completion, then `download_video` to retrieve the file.

## Model Tiers

### Image

| Tier | Model | Best For | Cost |
|------|-------|----------|------|
| nb2 (default) | `gemini-3.1-flash-image-preview` | Quick iterations, most tasks | ~$0.067/img |
| pro | `gemini-3-pro-image-preview` | Final renders, complex scenes | ~$0.134/img |

Both tiers support resolutions 1K, 2K, 4K and aspect ratios 1:1, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9.

### Video

| Tier | Model | Best For | Cost |
|------|-------|----------|------|
| lite (default) | `veo-3.1-lite-generate-preview` | High-volume, drafts | $0.05/sec (720p), $0.08/sec (1080p) |
| fast | `veo-3.1-fast-generate-preview` | Good quality iterations | $0.15/sec (720p/1080p), $0.35/sec (4k) |
| standard | `veo-3.1-generate-preview` | Final renders, 4K | $0.40/sec (720p/1080p), $0.60/sec (4k) |

Lite supports 720p and 1080p. Fast and Standard support 720p, 1080p, and 4K. Video extension (`extend_video`) is only available on Fast and Standard tiers.

You can pass the tier name (`lite`, `fast`, `standard`, `nb2`, `pro`) or a raw model ID directly.

## MCP Client Configuration

### Claude Code

Add to your Claude Code MCP settings (`~/.claude/settings.json` or project `.mcp.json`):

```json
{
  "mcpServers": {
    "gemini-media": {
      "command": "gemini-media-mcp",
      "env": {
        "GOOGLE_API_KEY": "your-api-key",
        "MEDIA_OUTPUT_DIR": "/path/to/output"
      }
    }
  }
}
```

Or if building from source:

```json
{
  "mcpServers": {
    "gemini-media": {
      "command": "/path/to/gemini-media-mcp",
      "env": {
        "GOOGLE_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Building from Source

```bash
git clone https://github.com/mordor-forge/gemini-media-mcp.git
cd gemini-media-mcp
go build ./cmd/gemini-media-mcp/
```

The binary will be created at `./gemini-media-mcp`.

To run tests:

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Make your changes and add tests
4. Run `go test ./...` and `go vet ./...`
5. Commit your changes
6. Open a pull request against `main`

## License

[Apache-2.0](LICENSE)
