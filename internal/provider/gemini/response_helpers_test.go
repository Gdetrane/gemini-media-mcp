package gemini

import (
	"testing"

	"google.golang.org/genai"
)

func TestExtractFirstNonEmptyInlineData_UsesLaterCandidate(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{InlineData: &genai.Blob{MIMEType: "image/png"}},
						{Text: "no media here"},
					},
				},
			},
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{InlineData: &genai.Blob{Data: []byte("gif-bytes"), MIMEType: "image/gif"}},
					},
				},
			},
		},
	}

	blob, ok := extractFirstNonEmptyInlineData(resp)
	if !ok {
		t.Fatal("expected to find inline data")
	}
	if string(blob.Data) != "gif-bytes" {
		t.Fatalf("blob.Data = %q, want %q", blob.Data, "gif-bytes")
	}
	if blob.MIMEType != "image/gif" {
		t.Fatalf("blob.MIMEType = %q, want %q", blob.MIMEType, "image/gif")
	}
}

func TestExtractMusicResponse_PrefersFirstAudioAndCollectsLyrics(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "[Verse] hello"},
						{InlineData: &genai.Blob{MIMEType: "audio/mpeg"}},
					},
				},
			},
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{InlineData: &genai.Blob{Data: []byte("first-audio"), MIMEType: "audio/mpeg"}},
						{Text: "[Chorus] world"},
						{InlineData: &genai.Blob{Data: []byte("second-audio"), MIMEType: "audio/mpeg"}},
					},
				},
			},
		},
	}

	audioData, mimeType, lyrics := extractMusicResponse(resp)
	if string(audioData) != "first-audio" {
		t.Fatalf("audioData = %q, want %q", audioData, "first-audio")
	}
	if mimeType != "audio/mpeg" {
		t.Fatalf("mimeType = %q, want %q", mimeType, "audio/mpeg")
	}
	if lyrics != "[Verse] hello\n[Chorus] world" {
		t.Fatalf("lyrics = %q, want %q", lyrics, "[Verse] hello\n[Chorus] world")
	}
}
