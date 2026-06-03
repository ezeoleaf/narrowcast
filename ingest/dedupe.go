package ingest

import (
	"net/url"
	"sort"
	"strings"
)

// DedupeByURL removes duplicate articles that share the same link.
// When two entries share a URL, the one with the newer Published time is kept;
// if neither has a date, the first seen wins. Order of first-seen URLs is preserved.
func DedupeByURL(articles []Article) []Article {
	if len(articles) == 0 {
		return nil
	}

	type slot struct {
		article Article
		order   int
	}

	byKey := make(map[string]slot)
	nextOrder := 0

	for _, a := range articles {
		key := articleKey(a)
		existing, ok := byKey[key]
		if !ok {
			byKey[key] = slot{article: a, order: nextOrder}
			nextOrder++
			continue
		}
		if preferArticle(a, existing.article) {
			byKey[key] = slot{article: a, order: existing.order}
		}
	}

	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return byKey[keys[i]].order < byKey[keys[j]].order
	})

	out := make([]Article, 0, len(keys))
	for _, k := range keys {
		out = append(out, byKey[k].article)
	}
	return out
}

// NormalizeURL canonicalizes a link for dedupe and seen-state keys.
func NormalizeURL(raw string) string {
	return normalizeURL(raw)
}

func articleKey(a Article) string {
	if key := normalizeURL(a.Link); key != "" {
		return key
	}
	return "no-url:" + a.Title + "\x00" + a.Source
}

func preferArticle(a, b Article) bool {
	if !a.Published.IsZero() && !b.Published.IsZero() {
		return a.Published.After(b.Published)
	}
	return !a.Published.IsZero()
}

func normalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return strings.TrimSuffix(raw, "/")
	}
	u.Fragment = ""
	u.RawFragment = ""
	return strings.TrimSuffix(u.String(), "/")
}
