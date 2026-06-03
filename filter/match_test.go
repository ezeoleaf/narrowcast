package filter

import (
	"testing"

	"github.com/ezeoleaf/narrowcast/ingest"
)

func TestWordBoundaryAvoidsSubstringFalsePositive(t *testing.T) {
	articles := []ingest.Article{
		{Title: "cargo ships delay", Summary: "trade"},
		{Title: "Why Go is fast", Summary: "golang"},
	}
	out, err := Apply(articles, Options{
		Include:    []string{"go"},
		MatchStyle: MatchWord,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Title != "Why Go is fast" {
		t.Fatalf("got %+v, want only Go article", out)
	}
}

func TestRegexLiteral(t *testing.T) {
	articles := []ingest.Article{
		{Title: "Raspberry Pi 5 review", Summary: ""},
		{Title: "Pi approximation", Summary: ""},
	}
	out, err := Apply(articles, Options{
		Include:    []string{`/(?i)raspberry\s+pi/`},
		MatchStyle: MatchWord,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("regex match got %d articles", len(out))
	}
}
