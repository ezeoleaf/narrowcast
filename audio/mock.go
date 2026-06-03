package audio

import (
	"context"
	"fmt"
	"io"
	"os"
)

// MockSpeaker prints what would be spoken.
type MockSpeaker struct {
	Out io.Writer
}

// NewMockSpeaker returns a mock that writes to stdout.
func NewMockSpeaker() *MockSpeaker {
	return &MockSpeaker{Out: os.Stdout}
}

func (m *MockSpeaker) out() io.Writer {
	if m.Out != nil {
		return m.Out
	}
	return os.Stdout
}

func (m *MockSpeaker) Speak(ctx context.Context, title, summary string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err := fmt.Fprintf(m.out(), "Speaking: %s - %s\n", title, summary)
	return err
}
