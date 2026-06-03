// Package broadcast runs one news cycle: ingest, filter, and speak.
package broadcast

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ezeoleaf/narrowcast/audio"
	"github.com/ezeoleaf/narrowcast/config"
	"github.com/ezeoleaf/narrowcast/filter"
	"github.com/ezeoleaf/narrowcast/ingest"
	"github.com/ezeoleaf/narrowcast/state"
)

// Result summarizes a single broadcast cycle.
type Result struct {
	Fetched int
	Matched int
	Queued  int
	Spoken  int
	Skipped int
}

// Cycle fetches feeds, filters, and speaks (or lists in dry-run mode).
func Cycle(ctx context.Context, cfg *config.Config, speaker audio.Speaker, store *state.SeenStore, fetchTimeout time.Duration, dryRun bool) (Result, error) {
	var res Result

	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	feeds := cfg.FeedSources()
	log.Printf("fetching %d feed(s)...", len(feeds))
	articles, _, err := ingest.FetchAll(fetchCtx, feeds)
	if err != nil {
		log.Printf("feed warnings: %v", err)
	}
	res.Fetched = len(articles)
	log.Printf("fetched %d article(s) total", res.Fetched)

	articles = ingest.DedupeByURL(articles)
	log.Printf("%d article(s) after dedupe", len(articles))

	maxAge, err := cfg.MaxAgeDuration()
	if err != nil {
		return res, err
	}
	curated, err := filter.Apply(articles, filter.Options{
		Include:    cfg.InterestTerms(),
		Exclude:    cfg.ExcludeTerms(),
		MatchAll:   cfg.MatchAll(),
		MatchStyle: cfg.MatchStyle,
		MaxAge:     maxAge,
	})
	if err != nil {
		return res, err
	}
	res.Matched = len(curated)
	log.Printf("%d article(s) match filters", res.Matched)

	if len(curated) == 0 {
		return res, nil
	}

	ingest.SortByPublishedNewest(curated)
	if cfg.MaxPerCycle > 0 && len(curated) > cfg.MaxPerCycle {
		curated = curated[:cfg.MaxPerCycle]
	}
	res.Queued = len(curated)

	pause, err := cfg.PauseBetweenDuration()
	if err != nil {
		return res, err
	}

	for n, article := range curated {
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		key := ingest.NormalizeURL(article.Link)
		if key != "" && store != nil && store.Has(key) {
			res.Skipped++
			continue
		}

		title := article.Title
		if title == "" {
			title = "(untitled)"
		}
		summary := article.Summary
		if summary == "" {
			summary = "No summary available."
		}

		res.Spoken++
		log.Printf("reading %d/%d from %s", res.Spoken, res.Queued, article.Source)

		if dryRun {
			log.Printf("[dry-run] %s — %s", title, truncateLog(summary, 120))
		} else if err := speaker.Speak(ctx, title, summary); err != nil {
			return res, fmt.Errorf("speak: %w", err)
		}

		if key != "" && store != nil {
			store.Mark(key)
		}

		if pause > 0 && n < len(curated)-1 {
			if err := sleepContext(ctx, pause); err != nil {
				return res, err
			}
		}
	}

	if store != nil {
		if err := store.Save(); err != nil {
			return res, fmt.Errorf("save state: %w", err)
		}
	}

	return res, nil
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func truncateLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
