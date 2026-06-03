// Package config loads user preferences from a YAML or JSON file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ezeoleaf/narrowcast/ingest"
	"gopkg.in/yaml.v3"
)

// Config holds Personal News Radio preferences.
type Config struct {
	Topics   []string `yaml:"topics" json:"topics"`
	Keywords []string `yaml:"keywords" json:"keywords"`
	Exclude  []string `yaml:"exclude" json:"exclude"`
	// MatchMode: "any" (default) or "all".
	MatchMode string `yaml:"match_mode" json:"match_mode"`
	// MatchStyle: "substring" (default) or "word" (word boundaries). Terms may be /regex/.
	MatchStyle   string `yaml:"match_style" json:"match_style"`
	MaxAge       string `yaml:"max_age" json:"max_age"`
	MaxPerCycle  int    `yaml:"max_per_cycle" json:"max_per_cycle"`
	PauseBetween string `yaml:"pause_between" json:"pause_between"`

	Feeds         FeedsList `yaml:"feeds" json:"feeds"`
	OPML          string    `yaml:"opml" json:"opml"`
	ResolvedFeeds []FeedEntry `yaml:"-" json:"-"`

	Schedule ScheduleConfig `yaml:"schedule" json:"schedule"`
	Audio    AudioConfig    `yaml:"audio" json:"audio"`
	State    StateConfig    `yaml:"state" json:"state"`
}

// ScheduleConfig controls daemon polling on a Raspberry Pi or server.
type ScheduleConfig struct {
	Interval string `yaml:"interval" json:"interval"`
}

// AudioConfig selects the TTS engine and its options.
type AudioConfig struct {
	Engine            string   `yaml:"engine" json:"engine"` // mock, espeak, piper, auto
	Fallback          []string `yaml:"fallback" json:"fallback"`
	MaxSummaryChars   int      `yaml:"max_summary_chars" json:"max_summary_chars"`
	MaxUtteranceChars int      `yaml:"max_utterance_chars" json:"max_utterance_chars"`
	Espeak            EspeakConfig `yaml:"espeak" json:"espeak"`
	Piper             PiperConfig  `yaml:"piper" json:"piper"`
}

// EspeakConfig configures espeak-ng / espeak.
type EspeakConfig struct {
	Binary string `yaml:"binary" json:"binary"`
	Voice  string `yaml:"voice" json:"voice"`
	Speed  int    `yaml:"speed" json:"speed"`
}

// PiperConfig configures Piper ONNX TTS plus WAV playback.
type PiperConfig struct {
	Binary     string `yaml:"binary" json:"binary"`
	Model      string `yaml:"model" json:"model"`
	PlayBinary string `yaml:"play_binary" json:"play_binary"`
}

// StateConfig persists URLs already read.
type StateConfig struct {
	File       string `yaml:"file" json:"file"`
	Persist    bool   `yaml:"persist" json:"persist"`
	TTL        string `yaml:"ttl" json:"ttl"`
	MaxEntries int    `yaml:"max_entries" json:"max_entries"`
}

// Load reads and parses a config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config extension %q (use .yaml, .yml, or .json)", ext)
	}

	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	resolved, err := cfg.ResolveFeeds()
	if err != nil {
		return nil, err
	}
	cfg.ResolvedFeeds = resolved
	return &cfg, nil
}

// FeedSources returns ingest sources from resolved feeds.
func (c *Config) FeedSources() []ingest.FeedSource {
	out := make([]ingest.FeedSource, 0, len(c.ResolvedFeeds))
	for _, f := range c.ResolvedFeeds {
		out = append(out, ingest.FeedSource{URL: f.URL, Label: f.Label})
	}
	return out
}

func (c *Config) applyDefaults() {
	if strings.TrimSpace(c.Audio.Engine) == "" {
		c.Audio.Engine = "mock"
	}
	if len(c.Audio.Fallback) == 0 {
		c.Audio.Fallback = []string{"piper", "espeak", "mock"}
	}
	if c.Audio.MaxSummaryChars <= 0 {
		c.Audio.MaxSummaryChars = 400
	}
	if c.Audio.MaxUtteranceChars <= 0 {
		c.Audio.MaxUtteranceChars = 2500
	}
	if c.State.File == "" {
		c.State.File = ".narrowcast/seen.json"
	}
	if c.State.MaxEntries <= 0 {
		c.State.MaxEntries = 2000
	}
	if strings.TrimSpace(c.State.TTL) == "" {
		c.State.TTL = "168h"
	}
	if c.MaxPerCycle <= 0 {
		c.MaxPerCycle = 10
	}
	if strings.TrimSpace(c.MatchMode) == "" {
		c.MatchMode = "any"
	}
	if strings.TrimSpace(c.MatchStyle) == "" {
		c.MatchStyle = "substring"
	}
}

func (c *Config) validate() error {
	if len(c.Topics) == 0 && len(c.Keywords) == 0 {
		return fmt.Errorf("config: at least one topic or keyword is required")
	}
	if _, err := c.ScheduleInterval(); err != nil {
		return err
	}
	if _, err := c.MaxAgeDuration(); err != nil {
		return err
	}
	if _, err := c.PauseBetweenDuration(); err != nil {
		return err
	}
	if _, err := c.StateTTLDuration(); err != nil {
		return err
	}
	mode := strings.ToLower(strings.TrimSpace(c.MatchMode))
	if mode != "any" && mode != "all" {
		return fmt.Errorf("config: match_mode must be \"any\" or \"all\", got %q", c.MatchMode)
	}
	c.MatchMode = mode
	style := strings.ToLower(strings.TrimSpace(c.MatchStyle))
	if style != "substring" && style != "word" {
		return fmt.Errorf("config: match_style must be \"substring\" or \"word\", got %q", c.MatchStyle)
	}
	c.MatchStyle = style
	engine := strings.ToLower(strings.TrimSpace(c.Audio.Engine))
	switch engine {
	case "mock", "espeak", "piper", "auto":
	default:
		return fmt.Errorf("config: audio.engine must be mock, espeak, piper, or auto, got %q", c.Audio.Engine)
	}
	c.Audio.Engine = engine
	if len(c.Feeds) == 0 && strings.TrimSpace(c.OPML) == "" {
		return fmt.Errorf("config: at least one feed URL or opml file is required")
	}
	return nil
}

// ScheduleInterval parses schedule.interval. Zero duration means run once.
func (c *Config) ScheduleInterval() (time.Duration, error) {
	return parseOptionalDuration(c.Schedule.Interval, "schedule.interval")
}

// MaxAgeDuration parses max_age. Zero means no age filter.
func (c *Config) MaxAgeDuration() (time.Duration, error) {
	return parseOptionalDuration(c.MaxAge, "max_age")
}

// PauseBetweenDuration parses pause_between.
func (c *Config) PauseBetweenDuration() (time.Duration, error) {
	return parseOptionalDuration(c.PauseBetween, "pause_between")
}

// StateTTLDuration parses state.ttl for pruning seen URLs.
func (c *Config) StateTTLDuration() (time.Duration, error) {
	s := strings.TrimSpace(c.State.TTL)
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("config: invalid state.ttl %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("config: state.ttl must be positive")
	}
	return d, nil
}

func parseOptionalDuration(s, field string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("config: invalid %s %q: %w", field, s, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("config: %s must be non-negative", field)
	}
	return d, nil
}

// MatchAll reports whether every include term must match.
func (c *Config) MatchAll() bool {
	return c.MatchMode == "all"
}

// InterestTerms returns topics and keywords for inclusion matching.
func (c *Config) InterestTerms() []string {
	out := make([]string, 0, len(c.Topics)+len(c.Keywords))
	for _, t := range c.Topics {
		if s := strings.TrimSpace(t); s != "" {
			out = append(out, s)
		}
	}
	for _, k := range c.Keywords {
		if s := strings.TrimSpace(k); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ExcludeTerms returns exclude keywords.
func (c *Config) ExcludeTerms() []string {
	out := make([]string, 0, len(c.Exclude))
	for _, e := range c.Exclude {
		if s := strings.TrimSpace(e); s != "" {
			out = append(out, s)
		}
	}
	return out
}
