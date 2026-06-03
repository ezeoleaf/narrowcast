package ingest

import "time"

// Article is a normalized headline from an RSS/Atom feed.
type Article struct {
	Title     string
	Summary   string
	Link      string
	Published time.Time
	Source    string // feed title or URL when title is missing
}
