// Package ingest fetches and parses RSS/Atom feeds.
package ingest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

const defaultFeedTimeout = 20 * time.Second

// FetchAll retrieves items from every feed concurrently (one parser per feed).
func FetchAll(ctx context.Context, feeds []FeedSource) ([]Article, []FeedHealth, error) {
	if len(feeds) == 0 {
		return nil, nil, nil
	}

	type result struct {
		health   FeedHealth
		articles []Article
	}

	results := make(chan result, len(feeds))
	var wg sync.WaitGroup

	for _, feed := range feeds {
		wg.Add(1)
		go func(src FeedSource) {
			defer wg.Done()
			start := time.Now()
			h := FeedHealth{
				Label: src.DisplayLabel(),
				URL:   src.URL,
			}
			parser := newParser()
			articles, err := fetchOne(ctx, parser, src)
			h.Elapsed = time.Since(start)
			h.Items = len(articles)
			h.Err = err
			results <- result{health: h, articles: articles}
		}(feed)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []Article
	var health []FeedHealth
	var errs []error
	for r := range results {
		logFeedHealth(r.health)
		health = append(health, r.health)
		if r.health.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.health.Label, r.health.Err))
			continue
		}
		all = append(all, r.articles...)
	}

	if len(errs) > 0 && len(all) == 0 {
		return nil, health, fmt.Errorf("all feeds failed: %w", joinErrors(errs))
	}
	return all, health, joinErrors(errs)
}

func logFeedHealth(h FeedHealth) {
	if h.Err != nil {
		log.Printf("feed %q: ERROR %v (%s)", h.Label, h.Err, h.Elapsed.Round(time.Millisecond))
		return
	}
	log.Printf("feed %q: ok %d items in %s", h.Label, h.Items, h.Elapsed.Round(time.Millisecond))
}

func newParser() *gofeed.Parser {
	p := gofeed.NewParser()
	p.UserAgent = "narrowcast/0.1"
	p.Client = &http.Client{Timeout: defaultFeedTimeout}
	return p
}

func fetchOne(ctx context.Context, parser *gofeed.Parser, src FeedSource) ([]Article, error) {
	feed, err := parser.ParseURLWithContext(src.URL, ctx)
	if err != nil {
		return nil, err
	}

	source := src.Label
	if source == "" {
		source = feed.Title
	}
	if source == "" {
		source = src.URL
	}

	articles := make([]Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		if item == nil {
			continue
		}
		articles = append(articles, Article{
			Title:     CleanText(strings.TrimSpace(item.Title)),
			Summary:   summaryText(item),
			Link:      firstLink(item),
			Published: itemPublished(item),
			Source:    source,
		})
	}
	return articles, nil
}

func summaryText(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	var raw string
	if s := strings.TrimSpace(item.Description); s != "" {
		raw = s
	} else if item.Content != "" {
		raw = strings.TrimSpace(item.Content)
	}
	if raw == "" {
		return ""
	}
	return SanitizeSummary(CleanText(raw))
}

func firstLink(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	if item.Link != "" {
		return item.Link
	}
	if len(item.Links) > 0 {
		return item.Links[0]
	}
	return ""
}

func itemPublished(item *gofeed.Item) time.Time {
	if item == nil {
		return time.Time{}
	}
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}
	return time.Time{}
}

// stripHTML removes simple tags for TTS-friendly text (lightweight, no deps).
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

type multiErr []error

func (e multiErr) Error() string {
	var parts []string
	for _, err := range e {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}
	return strings.Join(parts, "; ")
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return multiErr(errs)
}
