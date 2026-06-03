package ingest

import "testing"

func TestParseOPML(t *testing.T) {
	data := `<?xml version="1.0"?>
<opml version="2.0">
  <body>
    <outline text="Tech" title="Tech">
      <outline type="rss" text="HN" title="Hacker News" xmlUrl="https://hnrss.org/frontpage"/>
    </outline>
  </body>
</opml>`
	feeds, err := ParseOPML([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(feeds) != 1 {
		t.Fatalf("got %d feeds", len(feeds))
	}
	if feeds[0].URL != "https://hnrss.org/frontpage" {
		t.Fatalf("url=%q", feeds[0].URL)
	}
	if feeds[0].Label != "Hacker News" {
		t.Fatalf("label=%q", feeds[0].Label)
	}
}
