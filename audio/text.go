package audio

import (
	"strings"
	"unicode/utf8"
)

// FormatUtterance builds TTS-friendly text from a headline and summary.
func FormatUtterance(title, summary string, maxSummaryRunes, maxTotalRunes int) string {
	title = strings.TrimSpace(title)
	summary = strings.TrimSpace(summary)
	if title == "" {
		title = "untitled article"
	}
	if summary == "" {
		return truncateRunes(title, maxTotalRunes)
	}
	summary = truncateRunes(summary, maxSummaryRunes)
	utterance := title + ". " + summary
	return truncateRunes(utterance, maxTotalRunes)
}

func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	var n int
	for i := range s {
		if n == max {
			return strings.TrimSpace(s[:i]) + "…"
		}
		n++
	}
	return s
}
