package provider

// ImageRequest describes a text-to-image generation request.
type ImageRequest struct {
	Prompt      string `json:"prompt" jsonschema:"Text description of the image to generate"`
	Model       string `json:"model,omitempty" jsonschema:"Model tier: nb2 (default) or pro. Raw model IDs also accepted"`
	AspectRatio string `json:"aspectRatio,omitempty" jsonschema:"Aspect ratio (1:1, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9)"`
	Resolution  string `json:"resolution,omitempty" jsonschema:"Output resolution (1K, 2K, 4K)"`
}

// EditImageRequest describes an image editing request.
type EditImageRequest struct {
	Prompt    string `json:"prompt" jsonschema:"Text description of the edit to apply"`
	ImagePath string `json:"imagePath" jsonschema:"Path to the image to edit"`
	Model     string `json:"model,omitempty" jsonschema:"Model tier: nb2 (default) or pro"`
}

// ComposeRequest describes a multi-reference image composition request.
type ComposeRequest struct {
	Prompt          string   `json:"prompt" jsonschema:"Text description guiding the composition"`
	ReferenceImages []string `json:"referenceImages" jsonschema:"Paths to 1-3 reference images for style/content guidance"`
	Model           string   `json:"model,omitempty" jsonschema:"Model tier: nb2 (default) or pro"`
	AspectRatio     string   `json:"aspectRatio,omitempty" jsonschema:"Aspect ratio for the output"`
}

// VideoRequest describes a text-to-video generation request.
type VideoRequest struct {
	Prompt      string `json:"prompt" jsonschema:"Text description of the video to generate. Include audio cues for sound design"`
	Model       string `json:"model,omitempty" jsonschema:"Model tier: lite (default/cheapest), fast, or standard (highest quality). Raw model IDs also accepted"`
	AspectRatio string `json:"aspectRatio,omitempty" jsonschema:"Aspect ratio (16:9 or 9:16)"`
	Resolution  string `json:"resolution,omitempty" jsonschema:"Output resolution: 720p, 1080p, or 4k (lite supports 720p/1080p only)"`
	Duration    int    `json:"duration,omitempty" jsonschema:"Clip duration in seconds (4, 6, or 8)"`
}

// AnimateRequest describes an image-to-video generation request.
type AnimateRequest struct {
	Prompt      string `json:"prompt" jsonschema:"Text description guiding the animation"`
	ImagePath   string `json:"imagePath" jsonschema:"Path to image to use as the first frame"`
	Model       string `json:"model,omitempty" jsonschema:"Model tier: lite (default), fast, or standard"`
	AspectRatio string `json:"aspectRatio,omitempty" jsonschema:"Aspect ratio (16:9 or 9:16)"`
	Duration    int    `json:"duration,omitempty" jsonschema:"Clip duration in seconds (4, 6, or 8)"`
}

// ExtendRequest describes a video extension/chaining request.
type ExtendRequest struct {
	Prompt      string `json:"prompt" jsonschema:"Text description for the continuation"`
	OperationID string `json:"operationId" jsonschema:"Operation ID of the previous video generation"`
	Model       string `json:"model,omitempty" jsonschema:"Model tier (must match original). Standard and Fast only, Lite does not support extension"`
}

// ImageResult contains the result of an image generation operation.
type ImageResult struct {
	FilePath string `json:"filePath"`
	Model    string `json:"model"`
	MimeType string `json:"mimeType"`
}

// VideoOperation represents an in-progress async video generation.
type VideoOperation struct {
	OperationID string `json:"operationId"`
	Model       string `json:"model"`
}

// VideoStatus represents the current state of a video generation operation.
type VideoStatus struct {
	OperationID string `json:"operationId"`
	Done        bool   `json:"done"`
	Progress    string `json:"progress"` // "pending", "processing", "complete", "failed"
	Error       string `json:"error,omitempty"`
}

// VideoResult contains the result of a completed video generation.
type VideoResult struct {
	FilePath    string `json:"filePath"`
	OperationID string `json:"operationId"`
	Model       string `json:"model"`
	Duration    int    `json:"duration"`
}

// AudioRequest describes a text-to-speech audio generation request.
type AudioRequest struct {
	Prompt       string `json:"prompt" jsonschema:"Text to convert to speech or instructions for audio generation"`
	VoiceName    string `json:"voiceName,omitempty" jsonschema:"Prebuilt voice name (e.g. Aoede, Kore, Puck)"`
	LanguageCode string `json:"languageCode,omitempty" jsonschema:"Language code (e.g. en-US, it-IT, cs-CZ)"`
}

// AudioResult contains the result of an audio generation operation.
type AudioResult struct {
	FilePath string `json:"filePath"`
	Model    string `json:"model"`
	MimeType string `json:"mimeType"`
}

// MusicRequest describes a text-to-music generation request.
type MusicRequest struct {
	Prompt string `json:"prompt" jsonschema:"Text description of the music to generate. Supports genre, instruments, BPM, key, mood, structure tags like [Verse] [Chorus] [Bridge], and custom lyrics"`
	Model  string `json:"model,omitempty" jsonschema:"Model: clip (default, 30s clips) or full (up to 3 minutes, full songs with structure control)"`
}

// MusicResult contains the result of a music generation operation.
type MusicResult struct {
	FilePath string `json:"filePath"`
	Model    string `json:"model"`
	MimeType string `json:"mimeType"`
	Lyrics   string `json:"lyrics,omitempty"` // Generated lyrics/structure if returned
}

// ModelInfo describes an available model and its capabilities.
type ModelInfo struct {
	ID           string   `json:"id"`
	Tier         string   `json:"tier"`
	MediaType    string   `json:"mediaType"` // "image" or "video"
	Resolutions  []string `json:"resolutions"`
	AspectRatios []string `json:"aspectRatios"`
	PricePerSec  string   `json:"pricePerSec,omitempty"`
}
