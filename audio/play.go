package audio

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// defaultMediaPlayer returns a sensible audio player for MPEG/WAV files.
func defaultMediaPlayer() string {
	if runtime.GOOS == "darwin" {
		return "afplay"
	}
	return "mpv"
}

func playMedia(ctx context.Context, binary, path string) error {
	if binary == "" {
		binary = defaultMediaPlayer()
	}
	// mpv: play audio only, no terminal UI, exit when done.
	args := []string{path}
	if strings.Contains(binary, "mpv") || binary == "mpv" {
		args = []string{"--no-video", "--really-quiet", path}
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w (%s)", binary, err, strings.TrimSpace(string(out)))
	}
	return nil
}
