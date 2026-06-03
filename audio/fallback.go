package audio

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// FallbackSpeaker tries each Speaker in order until one succeeds.
type FallbackSpeaker struct {
	Names   []string
	Chain   []Speaker
}

func (f *FallbackSpeaker) Speak(ctx context.Context, title, summary string) error {
	var errs []error
	for i, s := range f.Chain {
		if err := s.Speak(ctx, title, summary); err == nil {
			return nil
		} else {
			name := "engine"
			if i < len(f.Names) {
				name = f.Names[i]
			}
			log.Printf("tts %s failed: %v", name, err)
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
	}
	return errors.Join(errs...)
}
