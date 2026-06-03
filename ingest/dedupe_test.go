package ingest

import (
	"testing"
	"time"
)

func TestDedupeByURL(t *testing.T) {
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	in := []Article{
		{Title: "A", Link: "https://example.com/post#comments", Published: older},
		{Title: "A newer", Link: "https://example.com/post/", Published: newer},
		{Title: "B", Link: "https://other.com/x"},
		{Title: "A duplicate", Link: "https://example.com/post", Published: older},
	}
	out := DedupeByURL(in)
	if len(out) != 2 {
		t.Fatalf("got %d articles, want 2: %+v", len(out), out)
	}
	if out[0].Title != "A newer" {
		t.Errorf("first = %q, want newer duplicate kept", out[0].Title)
	}
	if out[1].Title != "B" {
		t.Errorf("second = %q, want B", out[1].Title)
	}
}
