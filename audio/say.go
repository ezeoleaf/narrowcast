package audio

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// SaySpeaker uses the macOS built-in `say` command (no Homebrew deps).
type SaySpeaker struct {
	Binary            string
	Voice             string // e.g. Samantha, Alex — run `say -v ?`
	Rate              int    // words per minute (-r)
	MaxSummaryRunes   int
	MaxUtteranceRunes int
}

// AvailableOnDarwin reports whether the say engine can run on this machine.
func AvailableOnDarwin() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath("say")
	return err == nil
}

func (s *SaySpeaker) Speak(ctx context.Context, title, summary string) error {
	if !AvailableOnDarwin() {
		return fmt.Errorf("say: requires macOS with /usr/bin/say")
	}
	maxSummary := s.MaxSummaryRunes
	if maxSummary <= 0 {
		maxSummary = 400
	}
	maxTotal := s.MaxUtteranceRunes
	if maxTotal <= 0 {
		maxTotal = 2500
	}
	text := FormatUtterance(title, summary, maxSummary, maxTotal)

	tmp, err := os.CreateTemp("", "narrowcast-say-*.txt")
	if err != nil {
		return err
	}
	path := tmp.Name()
	if _, err := tmp.WriteString(text); err != nil {
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

	args := []string{"-f", path}
	if s.Voice != "" {
		args = append([]string{"-v", s.Voice}, args...)
	}
	if s.Rate > 0 {
		args = append(args, "-r", strconv.Itoa(s.Rate))
	}

	cmd := exec.CommandContext(ctx, s.binary(), args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("say: %w (%s)", err, trimExecOut(out))
	}
	return nil
}

func (s *SaySpeaker) binary() string {
	if s.Binary != "" {
		return s.Binary
	}
	return "say"
}

func trimExecOut(b []byte) string {
	const max = 200
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
