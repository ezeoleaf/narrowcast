package config

import "runtime"

func defaultAudioFallback() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"say", "piper", "espeak", "mock"}
	case "linux":
		return []string{"piper", "espeak", "mock"}
	default:
		return []string{"mock"}
	}
}
