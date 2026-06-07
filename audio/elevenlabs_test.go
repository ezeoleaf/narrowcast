package audio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestElevenLabsEngineSpeak(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("xi-api-key") != "test-key" {
			t.Fatalf("missing api key header")
		}
		if r.URL.Path != "/v1/text-to-speech/voice123" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte{0xFF, 0xFB}) // minimal mp3-ish bytes
	}))
	defer srv.Close()

	engine := &ElevenLabsEngine{
		APIKey:            "test-key",
		BaseURL:           srv.URL + "/v1",
		VoiceID:           "voice123",
		ModelID:           "eleven_turbo_v2_5",
		PlayBinary:        "true", // skip real playback in test — will fail play
		MaxSummaryRunes:   100,
		MaxUtteranceRunes: 200,
	}
	// Synthesize only — patch by testing synthesize directly
	data, err := engine.synthesize(context.Background(), "Hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 2 {
		t.Fatalf("expected audio bytes, got %d", len(data))
	}
}
