package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ezeoleaf/narrowcast/ingest"
	"gopkg.in/yaml.v3"
)

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// FeedEntry is a single RSS/Atom source with an optional display label.
type FeedEntry struct {
	URL   string `yaml:"url" json:"url"`
	Label string `yaml:"label" json:"label"`
}

// FeedsList accepts YAML/JSON feed entries as plain URLs or {url, label} maps.
type FeedsList []FeedEntry

// UnmarshalYAML supports "- https://..." and "- url: ...\n  label: ...".
func (f *FeedsList) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	var entries []FeedEntry
	switch node.Kind {
	case yaml.SequenceNode:
		for _, child := range node.Content {
			e, err := decodeFeedNode(child)
			if err != nil {
				return err
			}
			entries = append(entries, e)
		}
	case yaml.ScalarNode:
		e, err := decodeFeedNode(node)
		if err != nil {
			return err
		}
		entries = append(entries, e)
	default:
		return fmt.Errorf("feeds: expected list or string, got yaml kind %d", node.Kind)
	}
	*f = entries
	return nil
}

func decodeFeedNode(node *yaml.Node) (FeedEntry, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		url := strings.TrimSpace(node.Value)
		if url == "" {
			return FeedEntry{}, fmt.Errorf("feeds: empty URL")
		}
		return FeedEntry{URL: url}, nil
	case yaml.MappingNode:
		var e FeedEntry
		if err := node.Decode(&e); err != nil {
			return FeedEntry{}, err
		}
		e.URL = strings.TrimSpace(e.URL)
		e.Label = strings.TrimSpace(e.Label)
		if e.URL == "" {
			return FeedEntry{}, fmt.Errorf("feeds: entry missing url")
		}
		return e, nil
	default:
		return FeedEntry{}, fmt.Errorf("feeds: invalid entry kind %d", node.Kind)
	}
}

// UnmarshalJSON supports "https://..." strings or {"url","label"} objects.
func (f *FeedsList) UnmarshalJSON(data []byte) error {
	var raw []jsonFeed
	if err := jsonUnmarshal(data, &raw); err != nil {
		return err
	}
	entries := make([]FeedEntry, 0, len(raw))
	for _, item := range raw {
		e, err := item.feedEntry()
		if err != nil {
			return err
		}
		entries = append(entries, e)
	}
	*f = entries
	return nil
}

type jsonFeed struct {
	URL   string
	Label string
	plain string
}

func (j *jsonFeed) UnmarshalJSON(data []byte) error {
	var s string
	if err := jsonUnmarshal(data, &s); err == nil {
		j.plain = s
		return nil
	}
	type alias struct {
		URL   string `json:"url"`
		Label string `json:"label"`
	}
	var m alias
	if err := jsonUnmarshal(data, &m); err != nil {
		return err
	}
	j.URL = m.URL
	j.Label = m.Label
	return nil
}

func (j jsonFeed) feedEntry() (FeedEntry, error) {
	if j.plain != "" {
		url := strings.TrimSpace(j.plain)
		if url == "" {
			return FeedEntry{}, fmt.Errorf("feeds: empty URL")
		}
		return FeedEntry{URL: url}, nil
	}
	url := strings.TrimSpace(j.URL)
	if url == "" {
		return FeedEntry{}, fmt.Errorf("feeds: entry missing url")
	}
	return FeedEntry{URL: url, Label: strings.TrimSpace(j.Label)}, nil
}

// ResolveFeeds merges inline feeds with an optional OPML file (deduped by URL).
func (c *Config) ResolveFeeds() ([]FeedEntry, error) {
	seen := make(map[string]FeedEntry)
	add := func(e FeedEntry) {
		url := strings.TrimSpace(e.URL)
		if url == "" {
			return
		}
		e.URL = url
		e.Label = strings.TrimSpace(e.Label)
		if _, ok := seen[url]; !ok {
			seen[url] = e
		}
	}
	for _, e := range c.Feeds {
		add(e)
	}
	if path := strings.TrimSpace(c.OPML); path != "" {
		fromOPML, err := ingest.LoadOPML(path)
		if err != nil {
			return nil, fmt.Errorf("opml %q: %w", path, err)
		}
		for _, o := range fromOPML {
			add(FeedEntry{URL: o.URL, Label: o.Label})
		}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("config: at least one feed URL is required (feeds or opml)")
	}
	out := make([]FeedEntry, 0, len(seen))
	for _, e := range seen {
		out = append(out, e)
	}
	return out, nil
}
