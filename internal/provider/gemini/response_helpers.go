package gemini

import (
	"strings"

	"google.golang.org/genai"
)

func responseParts(resp *genai.GenerateContentResponse) []*genai.Part {
	if resp == nil {
		return nil
	}

	var parts []*genai.Part
	for _, candidate := range resp.Candidates {
		if candidate == nil || candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part != nil {
				parts = append(parts, part)
			}
		}
	}
	return parts
}

func extractFirstNonEmptyInlineData(resp *genai.GenerateContentResponse) (*genai.Blob, bool) {
	for _, part := range responseParts(resp) {
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			return part.InlineData, true
		}
	}
	return nil, false
}

func extractMusicResponse(resp *genai.GenerateContentResponse) ([]byte, string, string) {
	var audioData []byte
	var mimeType string
	var lyrics strings.Builder

	for _, part := range responseParts(resp) {
		if audioData == nil && part.InlineData != nil && len(part.InlineData.Data) > 0 {
			audioData = part.InlineData.Data
			mimeType = part.InlineData.MIMEType
		}
		if part.Text != "" {
			if lyrics.Len() > 0 {
				lyrics.WriteString("\n")
			}
			lyrics.WriteString(part.Text)
		}
	}

	return audioData, mimeType, lyrics.String()
}
