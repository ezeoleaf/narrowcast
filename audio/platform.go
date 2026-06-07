package audio

import "runtime"

// DefaultFallback returns a platform-appropriate TTS fallback chain for engine: auto.
func DefaultFallback() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"say", "elevenlabs", "piper", "espeak", "mock"}
	case "linux":
		return []string{"kokoro", "piper", "espeak", "mock"}
	default:
		return []string{"mock"}
	}
}

// DefaultEngine suggests the primary engine when none is set (optional hint for docs).
func DefaultEngine() string {
	if runtime.GOOS == "darwin" {
		return "auto"
	}
	return "mock"
}
