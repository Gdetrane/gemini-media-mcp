package gemini

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"
)

var (
	nowUTC               = func() time.Time { return time.Now().UTC() }
	randSource io.Reader = rand.Reader
)

// generateFilename produces a unique filename for a generated media file.
// Format: {mediaType}-{UTC timestamp}-{random hex}.{ext}
func generateFilename(mediaType, ext string) string {
	ts := nowUTC().Format("2006-01-02T15-04-05")
	id := shortID()
	return fmt.Sprintf("%s-%s-%s.%s", mediaType, ts, id, ext)
}

// shortID returns a 16-character random hex string for filename uniqueness.
// If cryptographic randomness is unavailable, it falls back to a timestamp-based suffix.
func shortID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(randSource, b); err != nil {
		return fmt.Sprintf("%016x", nowUTC().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}
