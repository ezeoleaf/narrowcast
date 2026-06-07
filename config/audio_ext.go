package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ElevenLabsConfig configures cloud TTS via ElevenLabs API.
type ElevenLabsConfig struct {
	// APIKey in config (prefer ELEVENLABS_API_KEY env via api_key_env).
	APIKey    string `yaml:"api_key" json:"api_key"`
	APIKeyEnv string `yaml:"api_key_env" json:"api_key_env"`
	VoiceID   string `yaml:"voice_id" json:"voice_id"`
	ModelID   string `yaml:"model_id" json:"model_id"`
	BaseURL   string `yaml:"base_url" json:"base_url"`
	PlayBinary string `yaml:"play_binary" json:"play_binary"` // afplay, mpv, mpg123
	Timeout   string `yaml:"timeout" json:"timeout"`
}

// KokoroConfig runs a local speak script (stdin = text).
type KokoroConfig struct {
	Script string   `yaml:"script" json:"script"`
	Args   []string `yaml:"args" json:"args"`
}

// TimeoutDuration parses elevenlabs.timeout (default 60s when engine runs).
func (e ElevenLabsConfig) TimeoutDuration() (time.Duration, error) {
	d, err := parseOptionalDuration(e.Timeout, "audio.elevenlabs.timeout")
	if err != nil {
		return 0, err
	}
	if d == 0 {
		return 60 * time.Second, nil
	}
	return d, nil
}

func (e ElevenLabsConfig) resolvedAPIKey() string {
	if k := strings.TrimSpace(e.APIKey); k != "" {
		return k
	}
	env := strings.TrimSpace(e.APIKeyEnv)
	if env == "" {
		env = "ELEVENLABS_API_KEY"
	}
	return strings.TrimSpace(os.Getenv(env))
}

func (c *Config) validateAudioEngine() error {
	engine := c.Audio.Engine
	switch engine {
	case "elevenlabs":
		if c.Audio.ElevenLabs.VoiceID == "" {
			return fmt.Errorf("config: audio.elevenlabs.voice_id is required when engine is elevenlabs")
		}
		if c.Audio.ElevenLabs.resolvedAPIKey() == "" {
			return fmt.Errorf("config: audio.elevenlabs.api_key or %s env is required",
				defaultElevenLabsEnv(c.Audio.ElevenLabs.APIKeyEnv))
		}
	case "kokoro":
		if strings.TrimSpace(c.Audio.Kokoro.Script) == "" {
			return fmt.Errorf("config: audio.kokoro.script is required when engine is kokoro")
		}
	}
	return nil
}

func defaultElevenLabsEnv(env string) string {
	if strings.TrimSpace(env) != "" {
		return env
	}
	return "ELEVENLABS_API_KEY"
}
