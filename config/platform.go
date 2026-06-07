package config

import "runtime"

func defaultAudioFallback() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"say", "elevenlabs", "piper", "espeak", "mock"}
	case "linux":
		return []string{"kokoro", "piper", "espeak", "mock"}
	default:
		return []string{"mock"}
	}
}
