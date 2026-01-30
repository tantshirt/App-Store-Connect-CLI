package progress

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Spinner renders a simple terminal spinner to a writer (intended for stderr).
// It does not emit anything unless started, and it never writes to stdout.
type Spinner struct {
	w      io.Writer
	msg    string
	frames []string
	period time.Duration

	doneOnce sync.Once
	done     chan struct{}
	wg       sync.WaitGroup

	mu      sync.Mutex
	lastLen int
}

// NewSpinner creates a spinner that renders "msg <frame>" on a single line.
func NewSpinner(w io.Writer, msg string) *Spinner {
	return &Spinner{
		w:      w,
		msg:    strings.TrimSpace(msg),
		frames: []string{"|", "/", "-", "\\"},
		period: 120 * time.Millisecond,
		done:   make(chan struct{}),
	}
}

// Start begins rendering until ctx is done or Stop is called.
func (s *Spinner) Start(ctx context.Context) {
	if s == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.period)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			case <-ticker.C:
				s.render(i)
				i++
			}
		}
	}()
}

// Stop stops rendering and clears the spinner line.
func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	s.doneOnce.Do(func() { close(s.done) })
	s.wg.Wait()
	s.clear()
}

func (s *Spinner) render(i int) {
	if len(s.frames) == 0 || s.w == nil {
		return
	}
	frame := s.frames[i%len(s.frames)]
	line := s.msg
	if line != "" {
		line = line + " "
	}
	line = line + frame

	// Use carriage return to stay on one line.
	_, _ = fmt.Fprintf(s.w, "\r%s", line)

	s.mu.Lock()
	s.lastLen = len(line)
	s.mu.Unlock()
}

func (s *Spinner) clear() {
	if s.w == nil {
		return
	}
	s.mu.Lock()
	n := s.lastLen
	s.lastLen = 0
	s.mu.Unlock()
	if n <= 0 {
		return
	}
	// Clear the previously rendered line and return carriage to column 0.
	_, _ = fmt.Fprintf(s.w, "\r%s\r", strings.Repeat(" ", n))
}

