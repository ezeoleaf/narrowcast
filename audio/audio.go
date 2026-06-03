// Package audio defines Text-to-Speech abstractions for reading headlines aloud.
package audio

import "context"

// Speaker turns article text into spoken output.
type Speaker interface {
	Speak(ctx context.Context, title, summary string) error
}
