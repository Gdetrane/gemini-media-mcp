package provider

import "context"

// ImageGenerator handles all image creation and manipulation operations.
type ImageGenerator interface {
	Generate(ctx context.Context, req ImageRequest) (*ImageResult, error)
	Edit(ctx context.Context, req EditImageRequest) (*ImageResult, error)
	Compose(ctx context.Context, req ComposeRequest) (*ImageResult, error)
}

// VideoGenerator handles video creation and lifecycle management.
type VideoGenerator interface {
	GenerateVideo(ctx context.Context, req VideoRequest) (*VideoOperation, error)
	AnimateImage(ctx context.Context, req AnimateRequest) (*VideoOperation, error)
	Extend(ctx context.Context, req ExtendRequest) (*VideoOperation, error)
	Status(ctx context.Context, operationID string) (*VideoStatus, error)
	Download(ctx context.Context, operationID string) (*VideoResult, error)
}

// AudioGenerator handles text-to-speech audio generation.
type AudioGenerator interface {
	GenerateAudio(ctx context.Context, req AudioRequest) (*AudioResult, error)
}

// MusicGenerator handles AI music generation.
type MusicGenerator interface {
	GenerateMusic(ctx context.Context, req MusicRequest) (*MusicResult, error)
}

// ModelLister provides discovery of available models and capabilities.
type ModelLister interface {
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
