package filter

import (
	"fmt"
	"regexp"
	"strings"
)

// MatchStyle controls how include/exclude terms are matched.
const (
	MatchSubstring = "substring"
	MatchWord      = "word"
)

// termMatcher builds a matcher for a single normalized (lowercase) term.
func termMatcher(term, style string) (func(string) bool, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return func(string) bool { return false }, nil
	}

	// Regex literal: /pattern/flags — only pattern between slashes, case-insensitive via (?i) if needed.
	if strings.HasPrefix(term, "/") && strings.Count(term, "/") >= 2 {
		last := strings.LastIndex(term, "/")
		if last > 0 {
			pattern := term[1:last]
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", term, err)
			}
			return func(haystack string) bool { return re.MatchString(haystack) }, nil
		}
	}

	switch strings.ToLower(strings.TrimSpace(style)) {
	case "", MatchSubstring:
		return func(haystack string) bool { return strings.Contains(haystack, term) }, nil
	case MatchWord:
		pattern := `(?i)\b` + regexp.QuoteMeta(term) + `\b`
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return func(haystack string) bool { return re.MatchString(haystack) }, nil
	default:
		return nil, fmt.Errorf("unknown match_style %q", style)
	}
}

func matchesAny(haystack string, terms []string, style string) (bool, error) {
	for _, term := range terms {
		m, err := termMatcher(term, style)
		if err != nil {
			return false, err
		}
		if m(haystack) {
			return true, nil
		}
	}
	return false, nil
}

func matchesAll(haystack string, terms []string, style string) (bool, error) {
	for _, term := range terms {
		m, err := termMatcher(term, style)
		if err != nil {
			return false, err
		}
		if !m(haystack) {
			return false, nil
		}
	}
	return true, nil
}
