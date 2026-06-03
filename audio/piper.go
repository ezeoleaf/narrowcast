package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PiperSpeaker runs Piper ONNX TTS and plays the WAV via aplay (or similar).
type PiperSpeaker struct {
	Binary            string
	Model             string
	PlayBinary        string
	MaxSummaryRunes   int
	MaxUtteranceRunes int
}

func (p *PiperSpeaker) Speak(ctx context.Context, title, summary string) error {
	if p.Model == "" {
		return fmt.Errorf("piper: model path is required")
	}
	maxSummary := p.MaxSummaryRunes
	if maxSummary <= 0 {
		maxSummary = 400
	}
	maxTotal := p.MaxUtteranceRunes
	if maxTotal <= 0 {
		maxTotal = 2500
	}
	text := FormatUtterance(title, summary, maxSummary, maxTotal)

	wav, err := os.CreateTemp("", "narrowcast-*.wav")
	if err != nil {
		return err
	}
	wavPath := wav.Name()
	_ = wav.Close()
	defer os.Remove(wavPath)

	synth := exec.CommandContext(ctx, p.binary(), "--model", p.Model, "--output_file", wavPath)
	synth.Stdin = strings.NewReader(text)
	if out, err := synth.CombinedOutput(); err != nil {
		return fmt.Errorf("piper: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	play := exec.CommandContext(ctx, p.playBinary(), wavPath)
	if out, err := play.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w (%s)", p.playBinary(), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (p *PiperSpeaker) binary() string {
	if p.Binary != "" {
		return p.Binary
	}
	return "piper"
}

func (p *PiperSpeaker) playBinary() string {
	if p.PlayBinary != "" {
		return p.PlayBinary
	}
	return "aplay"
}
