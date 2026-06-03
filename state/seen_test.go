package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSeenStoreTTLAndAtomicSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seen.json")

	s, err := Load(path, Options{TTL: time.Hour, MaxEntries: 100})
	if err != nil {
		t.Fatal(err)
	}
	s.Mark("https://example.com/old")
	s.mu.Lock()
	s.entries["https://example.com/old"] = time.Now().Add(-48 * time.Hour)
	s.mu.Unlock()
	s.Mark("https://example.com/new")
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s2, err := Load(path, Options{TTL: time.Hour, MaxEntries: 100})
	if err != nil {
		t.Fatal(err)
	}
	if s2.Has("https://example.com/old") {
		t.Fatal("old entry should be pruned on load")
	}
	if !s2.Has("https://example.com/new") {
		t.Fatal("new entry missing")
	}
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatal("state file empty")
	}
}
