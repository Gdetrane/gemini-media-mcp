//go:build e2e

package gemini

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mordor-forge/gemini-media-mcp/internal/provider"
)

// newE2EProvider creates a GeminiProvider for E2E tests. Skips the test
// if GOOGLE_API_KEY is not set.
func newE2EProvider(t *testing.T) *GeminiProvider {
	t.Helper()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}
	p, err := New(context.Background(), Config{
		APIKey:    apiKey,
		OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("creating provider: %v", err)
	}
	return p
}

// assertAudioFile checks that the file at path exists and has non-zero size.
// Returns the file size for further assertions.
func assertAudioFile(t *testing.T, path string) int64 {
	t.Helper()
	if path == "" {
		t.Fatal("file path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("file %q has zero size", path)
	}
	return info.Size()
}

// ---------------------------------------------------------------------------
// TTS (Text-to-Speech) Tests
// ---------------------------------------------------------------------------

func TestE2E_Audio_DefaultVoiceAndLanguage(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt: "The sun rose over the quiet village, casting long golden shadows across the cobblestone streets.",
	})
	if err != nil {
		t.Fatalf("GenerateAudio: %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if !strings.HasPrefix(result.MimeType, "audio/L16") {
		t.Errorf("expected PCM mime type (audio/L16*), got %q", result.MimeType)
	}
	t.Logf("path=%s model=%s mime=%s size=%d", result.FilePath, result.Model, result.MimeType, size)
}

func TestE2E_Audio_VoiceKore(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt:    "In a distant land, beyond the mountains, there lived a wise old owl who knew every secret of the forest.",
		VoiceName: "Kore",
	})
	if err != nil {
		t.Fatalf("GenerateAudio (Kore): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if !strings.HasPrefix(result.MimeType, "audio/L16") {
		t.Errorf("expected PCM mime type (audio/L16*), got %q", result.MimeType)
	}
	t.Logf("path=%s model=%s mime=%s size=%d", result.FilePath, result.Model, result.MimeType, size)
}

func TestE2E_Audio_VoicePuck(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt:    "The autumn leaves danced in the wind, swirling and tumbling through the crisp afternoon air.",
		VoiceName: "Puck",
	})
	if err != nil {
		t.Fatalf("GenerateAudio (Puck): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if !strings.HasPrefix(result.MimeType, "audio/L16") {
		t.Errorf("expected PCM mime type (audio/L16*), got %q", result.MimeType)
	}
	t.Logf("path=%s model=%s mime=%s size=%d", result.FilePath, result.Model, result.MimeType, size)
}

func TestE2E_Audio_LanguageItalian(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt:       "Il sole tramontava dietro le colline, dipingendo il cielo di arancione e rosa.",
		LanguageCode: "it-IT",
	})
	if err != nil {
		t.Fatalf("GenerateAudio (it-IT): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if !strings.HasPrefix(result.MimeType, "audio/L16") {
		t.Errorf("expected PCM mime type (audio/L16*), got %q", result.MimeType)
	}
	t.Logf("path=%s model=%s mime=%s size=%d", result.FilePath, result.Model, result.MimeType, size)
}

func TestE2E_Audio_LanguageCzech(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt:       "Podzimní listí padalo na zem a vítr šuměl mezi větvemi starých dubů.",
		LanguageCode: "cs-CZ",
	})
	if err != nil {
		t.Fatalf("GenerateAudio (cs-CZ): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if !strings.HasPrefix(result.MimeType, "audio/L16") {
		t.Errorf("expected PCM mime type (audio/L16*), got %q", result.MimeType)
	}
	t.Logf("path=%s model=%s mime=%s size=%d", result.FilePath, result.Model, result.MimeType, size)
}

func TestE2E_Audio_LongText(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	shortPrompt := "Good morning, world."
	shortResult, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt: shortPrompt,
	})
	if err != nil {
		t.Fatalf("GenerateAudio (short): %v", err)
	}
	shortSize := assertAudioFile(t, shortResult.FilePath)

	longPrompt := "The quick brown fox jumps over the lazy dog. " +
		"This sentence is used because it contains every letter of the English alphabet. " +
		"Text-to-speech systems must handle a wide range of phonemes, intonation patterns, " +
		"and prosodic features to produce natural-sounding speech. Long passages help verify " +
		"that the system maintains consistent quality and does not degrade over extended output."
	longResult, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt: longPrompt,
	})
	if err != nil {
		t.Fatalf("GenerateAudio (long): %v", err)
	}
	longSize := assertAudioFile(t, longResult.FilePath)

	if longSize <= shortSize {
		t.Errorf("long text (%d bytes) should be larger than short text (%d bytes)", longSize, shortSize)
	}
	t.Logf("short: path=%s size=%d | long: path=%s size=%d",
		shortResult.FilePath, shortSize, longResult.FilePath, longSize)
}

func TestE2E_Audio_EmptyPromptError(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	_, err := p.GenerateAudio(ctx, provider.AudioRequest{
		Prompt: "",
	})
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
	t.Logf("empty prompt error (expected): %v", err)
}

// ---------------------------------------------------------------------------
// Music (Lyria) Tests
// ---------------------------------------------------------------------------

func TestE2E_Music_ClipSimpleInstrumental(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "A gentle acoustic guitar in C major, 90 BPM",
		Model:  "clip",
	})
	if err != nil {
		t.Fatalf("GenerateMusic (clip, instrumental): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if result.MimeType != "audio/mpeg" {
		t.Errorf("expected mime audio/mpeg, got %q", result.MimeType)
	}
	lyricsSnippet := truncate(result.Lyrics, 100)
	t.Logf("path=%s model=%s mime=%s size=%d lyrics=%q",
		result.FilePath, result.Model, result.MimeType, size, lyricsSnippet)
}

func TestE2E_Music_ClipElectronic(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "Upbeat electronic dance music, 128 BPM, energetic synths",
		Model:  "clip",
	})
	if err != nil {
		t.Fatalf("GenerateMusic (clip, electronic): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if result.MimeType != "audio/mpeg" {
		t.Errorf("expected mime audio/mpeg, got %q", result.MimeType)
	}
	lyricsSnippet := truncate(result.Lyrics, 100)
	t.Logf("path=%s model=%s mime=%s size=%d lyrics=%q",
		result.FilePath, result.Model, result.MimeType, size, lyricsSnippet)
}

func TestE2E_Music_ClipStructureTags(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "[Intro] Soft piano [Verse] Add drums and bass [Chorus] Full band, uplifting",
		Model:  "clip",
	})
	if err != nil {
		t.Fatalf("GenerateMusic (clip, structure tags): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if result.MimeType != "audio/mpeg" {
		t.Errorf("expected mime audio/mpeg, got %q", result.MimeType)
	}
	lyricsSnippet := truncate(result.Lyrics, 100)
	t.Logf("path=%s model=%s mime=%s size=%d lyrics=%q",
		result.FilePath, result.Model, result.MimeType, size, lyricsSnippet)
}

func TestE2E_Music_FullModel(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	result, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "A short jazz piano solo, mellow and smooth, 80 BPM",
		Model:  "full",
	})
	if err != nil {
		t.Fatalf("GenerateMusic (full): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if result.MimeType != "audio/mpeg" {
		t.Errorf("expected mime audio/mpeg, got %q", result.MimeType)
	}
	lyricsSnippet := truncate(result.Lyrics, 100)
	t.Logf("path=%s model=%s mime=%s size=%d lyrics=%q",
		result.FilePath, result.Model, result.MimeType, size, lyricsSnippet)
}

func TestE2E_Music_LyricsCaptured(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	// Use a prompt with explicit lyrics to maximize the chance of text return.
	result, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "[Verse] Walking through the morning light, coffee in my hand [Chorus] This is the good life, just like we planned",
		Model:  "full",
	})
	if err != nil {
		t.Fatalf("GenerateMusic (lyrics capture): %v", err)
	}

	size := assertAudioFile(t, result.FilePath)
	if result.Lyrics == "" {
		t.Log("WARNING: Lyrics field is empty - model did not return text. This may be expected for some prompts.")
	} else {
		t.Logf("lyrics captured (%d chars)", len(result.Lyrics))
	}
	lyricsSnippet := truncate(result.Lyrics, 100)
	t.Logf("path=%s model=%s mime=%s size=%d lyrics=%q",
		result.FilePath, result.Model, result.MimeType, size, lyricsSnippet)
}

func TestE2E_Music_EmptyPromptError(t *testing.T) {
	p := newE2EProvider(t)
	ctx := context.Background()

	_, err := p.GenerateMusic(ctx, provider.MusicRequest{
		Prompt: "",
		Model:  "clip",
	})
	if err == nil {
		t.Fatal("expected error for empty prompt, got nil")
	}
	t.Logf("empty prompt error (expected): %v", err)
}

// truncate returns the first n characters of s, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
