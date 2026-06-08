package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultElevenLabsBaseURL = "https://api.elevenlabs.io/v1"
	defaultElevenLabsModel   = "eleven_turbo_v2_5"
	defaultElevenLabsKeyEnv  = "ELEVENLABS_API_KEY"
)

// ElevenLabsEngine synthesizes speech via the ElevenLabs cloud API.
type ElevenLabsEngine struct {
	APIKey            string
	BaseURL           string
	VoiceID           string
	ModelID           string
	PlayBinary        string
	Timeout           time.Duration
	MaxSummaryRunes   int
	MaxUtteranceRunes int
	Client            *http.Client
}

type elevenLabsRequest struct {
	Text     string `json:"text"`
	ModelID  string `json:"model_id"`
	Language string `json:"language_code,omitempty"`
}

func (e *ElevenLabsEngine) Speak(ctx context.Context, title, summary string) error {
	if e.APIKey == "" {
		return fmt.Errorf("elevenlabs: api key is required (config or %s)", defaultElevenLabsKeyEnv)
	}
	if e.VoiceID == "" {
		return fmt.Errorf("elevenlabs: voice_id is required")
	}

	maxSummary := e.MaxSummaryRunes
	if maxSummary <= 0 {
		maxSummary = 400
	}
	maxTotal := e.MaxUtteranceRunes
	if maxTotal <= 0 {
		maxTotal = 2500
	}
	text := FormatUtterance(title, summary, maxSummary, maxTotal)

	audioData, err := e.synthesize(ctx, text)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "narrowcast-elevenlabs-*.mp3")
	if err != nil {
		return err
	}
	path := tmp.Name()
	if _, err := tmp.Write(audioData); err != nil {
		_ = tmp.Close()
		_ = os.Remove(path)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			log.Printf("error removing temp file: %v", err)
		}
	}()

	return playMedia(ctx, e.PlayBinary, path)
}

func (e *ElevenLabsEngine) synthesize(ctx context.Context, text string) ([]byte, error) {
	base := e.BaseURL
	if base == "" {
		base = defaultElevenLabsBaseURL
	}
	base = strings.TrimRight(base, "/")
	url := base + "/text-to-speech/" + e.VoiceID

	model := e.ModelID
	if model == "" {
		model = defaultElevenLabsModel
	}
	body, err := json.Marshal(elevenLabsRequest{Text: text, ModelID: model})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", e.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	client := e.Client
	if client == nil {
		timeout := e.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("elevenlabs: HTTP %d: %s", resp.StatusCode, trimExecOut(data))
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("elevenlabs: empty audio response")
	}
	return data, nil
}

// ResolveElevenLabsAPIKey returns the configured key or reads from env.
func ResolveElevenLabsAPIKey(cfgKey, envName string) string {
	if k := strings.TrimSpace(cfgKey); k != "" {
		return k
	}
	if envName == "" {
		envName = defaultElevenLabsKeyEnv
	}
	return strings.TrimSpace(os.Getenv(envName))
}
