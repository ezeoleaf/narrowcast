// Package filter keeps articles that match the user's topics and keywords.
package filter

import (
	"strings"
	"time"

	"github.com/ezeoleaf/narrowcast/ingest"
)

// Options controls inclusion, exclusion, age, and match semantics.
type Options struct {
	Include    []string
	Exclude    []string
	MatchAll   bool
	MatchStyle string // substring (default) or word
	MaxAge     time.Duration
}

// Apply returns articles passing include/exclude rules and max age.
func Apply(articles []ingest.Article, opts Options) ([]ingest.Article, error) {
	include := normalizeTerms(opts.Include)
	if len(include) == 0 {
		return nil, nil
	}
	exclude := normalizeTerms(opts.Exclude)
	style := opts.MatchStyle
	if style == "" {
		style = MatchSubstring
	}
	cutoff := time.Time{}
	if opts.MaxAge > 0 {
		cutoff = time.Now().Add(-opts.MaxAge)
	}

	matched := make([]ingest.Article, 0, len(articles))
	for _, a := range articles {
		if opts.MaxAge > 0 && !a.Published.IsZero() && a.Published.Before(cutoff) {
			continue
		}
		haystack := strings.ToLower(a.Title + " " + a.Summary)
		if len(exclude) > 0 {
			ok, err := matchesAny(haystack, exclude, style)
			if err != nil {
				return nil, err
			}
			if ok {
				continue
			}
		}
		var ok bool
		var err error
		if opts.MatchAll {
			ok, err = matchesAll(haystack, include, style)
		} else {
			ok, err = matchesAny(haystack, include, style)
		}
		if err != nil {
			return nil, err
		}
		if ok {
			matched = append(matched, a)
		}
	}
	return matched, nil
}

// ByInterest is a convenience wrapper for include-only any-match filtering.
func ByInterest(articles []ingest.Article, terms []string) ([]ingest.Article, error) {
	return Apply(articles, Options{Include: terms})
}

func normalizeTerms(terms []string) []string {
	out := make([]string, 0, len(terms))
	for _, t := range terms {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		// Preserve /regex/ literals; lowercase plain terms for substring/word match.
		if strings.HasPrefix(t, "/") && strings.Count(t, "/") >= 2 {
			out = append(out, t)
		} else {
			out = append(out, strings.ToLower(t))
		}
	}
	return out
}
