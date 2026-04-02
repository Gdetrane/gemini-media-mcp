package gemini

import (
	"crypto/rand"
	"fmt"
	"time"
)

// generateFilename produces a unique filename for a generated media file.
// Format: {mediaType}-{UTC timestamp}-{random hex}.{ext}
func generateFilename(mediaType, ext string) string {
	ts := time.Now().UTC().Format("2006-01-02T15-04-05")
	id := shortID()
	return fmt.Sprintf("%s-%s-%s.%s", mediaType, ts, id, ext)
}

// shortID returns a 4-character random hex string for filename uniqueness.
func shortID() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
