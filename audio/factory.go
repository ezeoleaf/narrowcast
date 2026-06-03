package audio

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ezeoleaf/narrowcast/config"
)

// NewSpeaker builds a Speaker from config (single engine or fallback chain).
// The returned string describes what will be used (for logs).
func NewSpeaker(cfg config.AudioConfig) (Speaker, string, error) {
	engine := strings.ToLower(strings.TrimSpace(cfg.Engine))
	if engine == "auto" {
		return newFallbackChain(cfg, cfg.Fallback)
	}
	if len(cfg.Fallback) > 0 && engine != "mock" {
		seen := map[string]bool{engine: true}
		ordered := []string{engine}
		for _, n := range cfg.Fallback {
			n = strings.ToLower(strings.TrimSpace(n))
			if n == "" || seen[n] {
				continue
			}
			seen[n] = true
			ordered = append(ordered, n)
		}
		return newFallbackChain(cfg, ordered)
	}
	s, err := buildSpeaker(cfg, engine)
	if err != nil {
		return nil, "", err
	}
	return s, engine, nil
}

func newFallbackChain(cfg config.AudioConfig, names []string) (Speaker, string, error) {
	var chain []Speaker
	var used []string
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}
		s, err := buildSpeaker(cfg, name)
		if err != nil {
			if errors.Is(err, errSkipEngine) {
				continue
			}
			return nil, "", err
		}
		chain = append(chain, s)
		used = append(used, name)
	}
	if len(chain) == 0 {
		return NewMockSpeaker(), "mock (fallback empty)", nil
	}
	if len(chain) == 1 {
		return chain[0], used[0], nil
	}
	return &FallbackSpeaker{Names: used, Chain: chain}, strings.Join(used, " → "), nil
}

var errSkipEngine = fmt.Errorf("engine not configured")

func buildSpeaker(cfg config.AudioConfig, engine string) (Speaker, error) {
	switch engine {
	case "mock":
		return NewMockSpeaker(), nil
	case "espeak":
		return &EspeakSpeaker{
			Binary:            cfg.Espeak.Binary,
			Voice:             cfg.Espeak.Voice,
			Speed:             cfg.Espeak.Speed,
			MaxSummaryRunes:   cfg.MaxSummaryChars,
			MaxUtteranceRunes: cfg.MaxUtteranceChars,
		}, nil
	case "piper":
		if cfg.Piper.Model == "" {
			return nil, errSkipEngine
		}
		return &PiperSpeaker{
			Binary:            cfg.Piper.Binary,
			Model:             cfg.Piper.Model,
			PlayBinary:        cfg.Piper.PlayBinary,
			MaxSummaryRunes:   cfg.MaxSummaryChars,
			MaxUtteranceRunes: cfg.MaxUtteranceChars,
		}, nil
	default:
		return nil, fmt.Errorf("unknown audio engine %q", engine)
	}
}
