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
		names := cfg.Fallback
		if len(names) == 0 {
			names = DefaultFallback()
		}
		return newFallbackChain(cfg, names)
	}
	if len(cfg.Fallback) > 0 && supportsFallback(engine) {
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

func supportsFallback(engine string) bool {
	switch engine {
	case "piper", "espeak", "elevenlabs", "kokoro":
		return true
	default:
		return false
	}
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
		return NewMockSpeaker(), "mock (no TTS engines available)", nil
	}
	if len(chain) == 1 {
		return chain[0], used[0], nil
	}
	return &FallbackSpeaker{Names: used, Chain: chain}, strings.Join(used, " → "), nil
}

var errSkipEngine = fmt.Errorf("engine not configured")

func buildSpeaker(cfg config.AudioConfig, engine string) (Speaker, error) {
	maxSummary := cfg.MaxSummaryChars
	maxUtterance := cfg.MaxUtteranceChars

	switch engine {
	case "mock":
		return NewMockSpeaker(), nil
	case "say":
		if !AvailableOnDarwin() {
			return nil, errSkipEngine
		}
		return &SaySpeaker{
			Binary:            cfg.Say.Binary,
			Voice:             cfg.Say.Voice,
			Rate:              cfg.Say.Rate,
			MaxSummaryRunes:   maxSummary,
			MaxUtteranceRunes: maxUtterance,
		}, nil
	case "espeak":
		return &EspeakSpeaker{
			Binary:            cfg.Espeak.Binary,
			Voice:             cfg.Espeak.Voice,
			Speed:             cfg.Espeak.Speed,
			MaxSummaryRunes:   maxSummary,
			MaxUtteranceRunes: maxUtterance,
		}, nil
	case "piper":
		if cfg.Piper.Model == "" {
			return nil, errSkipEngine
		}
		return &PiperSpeaker{
			Binary:            cfg.Piper.Binary,
			Model:             cfg.Piper.Model,
			PlayBinary:        cfg.Piper.PlayBinary,
			MaxSummaryRunes:   maxSummary,
			MaxUtteranceRunes: maxUtterance,
		}, nil
	case "elevenlabs":
		key := ResolveElevenLabsAPIKey(cfg.ElevenLabs.APIKey, cfg.ElevenLabs.APIKeyEnv)
		if key == "" || cfg.ElevenLabs.VoiceID == "" {
			return nil, errSkipEngine
		}
		timeout, err := cfg.ElevenLabs.TimeoutDuration()
		if err != nil {
			return nil, err
		}
		return &ElevenLabsEngine{
			APIKey:            key,
			BaseURL:           cfg.ElevenLabs.BaseURL,
			VoiceID:           cfg.ElevenLabs.VoiceID,
			ModelID:           cfg.ElevenLabs.ModelID,
			PlayBinary:        cfg.ElevenLabs.PlayBinary,
			Timeout:           timeout,
			MaxSummaryRunes:   maxSummary,
			MaxUtteranceRunes: maxUtterance,
		}, nil
	case "kokoro":
		if !KokoroScriptReady(cfg.Kokoro.Script) {
			return nil, errSkipEngine
		}
		return &KokoroEngine{
			Script:            cfg.Kokoro.Script,
			Args:              cfg.Kokoro.Args,
			MaxSummaryRunes:   maxSummary,
			MaxUtteranceRunes: maxUtterance,
		}, nil
	default:
		return nil, fmt.Errorf("unknown audio engine %q", engine)
	}
}
