package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// KokoroEngine runs a local Kokoro (or compatible) speak script on the Pi/host.
// Script contract: read utterance text from stdin; synthesize and play audio;
// exit 0 on success.
type KokoroEngine struct {
	Script            string
	Args              []string
	MaxSummaryRunes   int
	MaxUtteranceRunes int
}

func (k *KokoroEngine) Speak(ctx context.Context, title, summary string) error {
	script := strings.TrimSpace(k.Script)
	if script == "" {
		return fmt.Errorf("kokoro: script path is required")
	}
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("kokoro: script %q: %w", script, err)
	}

	maxSummary := k.MaxSummaryRunes
	if maxSummary <= 0 {
		maxSummary = 400
	}
	maxTotal := k.MaxUtteranceRunes
	if maxTotal <= 0 {
		maxTotal = 2500
	}
	text := FormatUtterance(title, summary, maxSummary, maxTotal)

	args := append([]string{}, k.Args...)
	cmd := exec.CommandContext(ctx, script, args...)
	cmd.Stdin = strings.NewReader(text)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kokoro: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// KokoroScriptReady reports whether the configured script exists.
func KokoroScriptReady(script string) bool {
	script = strings.TrimSpace(script)
	if script == "" {
		return false
	}
	_, err := os.Stat(script)
	return err == nil
}
