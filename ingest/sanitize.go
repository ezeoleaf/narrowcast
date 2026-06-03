package ingest

import (
	"regexp"
	"strings"
)

var (
	reSubmittedBy = regexp.MustCompile(`(?i)\s*submitted by\s+.*$`)
	reLinkTag     = regexp.MustCompile(`(?i)\s*\[link\]\s*(\[comments\])?\s*$`)
	reHNMeta      = regexp.MustCompile(`(?i)\s*(article url:|comments url:|points:|#\s*comments:)\s*https?://\S+`)
	reMultiSpace  = regexp.MustCompile(`\s{2,}`)
)

// SanitizeSummary strips common RSS boilerplate for TTS.
func SanitizeSummary(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = reHNMeta.ReplaceAllString(s, "")
	s = reSubmittedBy.ReplaceAllString(s, "")
	s = reLinkTag.ReplaceAllString(s, "")
	s = reMultiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
