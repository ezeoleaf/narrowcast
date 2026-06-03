package ingest

import (
	"html"
	"strings"
)

// CleanText prepares feed text for display and TTS: strips tags, unescapes HTML entities.
func CleanText(s string) string {
	s = stripHTML(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
