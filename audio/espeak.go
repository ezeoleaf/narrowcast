package audio

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// EspeakSpeaker uses espeak-ng (or espeak) for local TTS.
type EspeakSpeaker struct {
	Binary            string
	Voice             string
	Speed             int
	MaxSummaryRunes   int
	MaxUtteranceRunes int
}

func (e *EspeakSpeaker) Speak(ctx context.Context, title, summary string) error {
	maxSummary := e.MaxSummaryRunes
	if maxSummary <= 0 {
		maxSummary = 400
	}
	maxTotal := e.MaxUtteranceRunes
	if maxTotal <= 0 {
		maxTotal = 2500
	}
	text := FormatUtterance(title, summary, maxSummary, maxTotal)

	args := []string{}
	if e.Voice != "" {
		args = append(args, "-v", e.Voice)
	}
	if e.Speed > 0 {
		args = append(args, "-s", strconv.Itoa(e.Speed))
	}
	args = append(args, text)

	cmd := exec.CommandContext(ctx, e.binary(), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("espeak: %w", err)
	}
	return nil
}

func (e *EspeakSpeaker) binary() string {
	if e.Binary != "" {
		return e.Binary
	}
	return "espeak-ng"
}
