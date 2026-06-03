// Package state tracks article URLs already read aloud.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type diskFormat struct {
	Entries map[string]time.Time `json:"entries"`
}

// SeenStore remembers article links that were already spoken.
type SeenStore struct {
	mu         sync.Mutex
	path       string
	entries    map[string]time.Time
	dirty      bool
	ttl        time.Duration
	maxEntries int
}

// Options configures pruning and persistence.
type Options struct {
	TTL        time.Duration
	MaxEntries int
}

// Load opens or creates a seen-URL store at path.
func Load(path string, opts Options) (*SeenStore, error) {
	s := &SeenStore{
		path:       path,
		entries:    make(map[string]time.Time),
		ttl:        opts.TTL,
		maxEntries: opts.MaxEntries,
	}
	if path == "" {
		return s, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	if err := s.decode(data); err != nil {
		return nil, err
	}
	s.pruneLocked(time.Now())
	return s, nil
}

func (s *SeenStore) decode(data []byte) error {
	var probe struct {
		Entries json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(data, &probe); err == nil && len(probe.Entries) > 0 {
		var f diskFormat
		if err := json.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("parse state: %w", err)
		}
		if f.Entries != nil {
			s.entries = f.Entries
		}
		return nil
	}
	var legacy []string
	if err := json.Unmarshal(data, &legacy); err == nil && len(legacy) > 0 {
		now := time.Now().UTC()
		for _, u := range legacy {
			if u != "" {
				s.entries[u] = now
			}
		}
		s.dirty = true
		return nil
	}
	return nil
}

// Has reports whether url was already spoken.
func (s *SeenStore) Has(url string) bool {
	if s == nil || url == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.entries[url]
	return ok
}

// Mark records url as spoken at now (UTC).
func (s *SeenStore) Mark(url string) {
	if s == nil || url == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[url]; ok {
		return
	}
	s.entries[url] = time.Now().UTC()
	s.dirty = true
}

// Save persists the store atomically if it changed.
func (s *SeenStore) Save() error {
	if s == nil || s.path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(time.Now())
	if !s.dirty {
		return nil
	}
	f := diskFormat{Entries: s.entries}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(s.path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir state dir: %w", err)
		}
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write state temp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename state: %w", err)
	}
	s.dirty = false
	return nil
}

func (s *SeenStore) pruneLocked(now time.Time) {
	if s.ttl > 0 {
		for u, at := range s.entries {
			if now.Sub(at) > s.ttl {
				delete(s.entries, u)
				s.dirty = true
			}
		}
	}
	if s.maxEntries > 0 && len(s.entries) > s.maxEntries {
		type pair struct {
			url string
			at  time.Time
		}
		all := make([]pair, 0, len(s.entries))
		for u, at := range s.entries {
			all = append(all, pair{u, at})
		}
		sort.Slice(all, func(i, j int) bool {
			return all[i].at.Before(all[j].at)
		})
		remove := len(all) - s.maxEntries
		for i := 0; i < remove; i++ {
			delete(s.entries, all[i].url)
		}
		s.dirty = true
	}
}
