package filter

import (
	"testing"
	"time"

	"github.com/ezeoleaf/narrowcast/ingest"
)

func TestApplyExcludeAndMatchAll(t *testing.T) {
	now := time.Now()
	old := now.Add(-72 * time.Hour)

	articles := []ingest.Article{
		{Title: "Open source Pi project", Summary: "tech", Published: now},
		{Title: "Sports crypto roundup", Summary: "markets", Published: now},
		{Title: "Climate science report", Summary: "open source data", Published: old},
	}

	out, err := Apply(articles, Options{
		Include:  []string{"open source"},
		Exclude:  []string{"crypto"},
		MatchAll: true,
		MaxAge:   48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Title != "Open source Pi project" {
		t.Fatalf("expected Pi project only, got %+v", out)
	}

	out, err = Apply(articles, Options{
		Include:  []string{"open", "source"},
		MatchAll: true,
		MaxAge:   48 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Title != "Open source Pi project" {
		t.Fatalf("unexpected match: %+v", out)
	}
}
