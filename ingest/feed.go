package ingest

import "time"

// FeedSource is a feed URL with an optional display label for logs and article.Source.
type FeedSource struct {
	URL   string
	Label string
}

// DisplayLabel returns the configured label or a short URL fallback.
func (f FeedSource) DisplayLabel() string {
	if f.Label != "" {
		return f.Label
	}
	return f.URL
}

// FeedHealth summarizes one feed fetch attempt.
type FeedHealth struct {
	Label   string
	URL     string
	Items   int
	Elapsed time.Duration
	Err     error
}
