package ingest

import "sort"

// SortByPublishedNewest orders articles with newest Published first.
// Articles without a date sort after dated ones.
func SortByPublishedNewest(articles []Article) {
	sort.SliceStable(articles, func(i, j int) bool {
		pi, pj := articles[i].Published, articles[j].Published
		if pi.IsZero() && pj.IsZero() {
			return false
		}
		if pi.IsZero() {
			return false
		}
		if pj.IsZero() {
			return true
		}
		return pi.After(pj)
	})
}
