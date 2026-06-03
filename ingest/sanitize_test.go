package ingest

import (
	"strings"
	"testing"
)

func TestSanitizeSummary(t *testing.T) {
	in := "Great post. Article URL: https://x.com/a Points: 42 # Comments: 7 submitted by /u/foo [link] [comments]"
	got := SanitizeSummary(in)
	if strings.Contains(got, "Article URL") || strings.Contains(got, "submitted by") {
		t.Fatalf("boilerplate remained: %q", got)
	}
	if !strings.Contains(got, "Great post") {
		t.Fatalf("content lost: %q", got)
	}
}
