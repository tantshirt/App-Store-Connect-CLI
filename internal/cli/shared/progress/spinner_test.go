package progress

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestSpinner_StartStopWritesAndClears(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner(&buf, "Working")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	s.period = 1 * time.Millisecond
	s.Start(ctx)

	// Let it tick at least once.
	time.Sleep(3 * time.Millisecond)
	s.Stop()

	out := buf.String()
	if out == "" {
		t.Fatal("expected spinner output, got empty")
	}
	// Should have rendered and cleared using carriage returns.
	if !bytes.Contains([]byte(out), []byte("\rWorking")) {
		t.Fatalf("expected rendered line, got %q", out)
	}
	if !bytes.Contains([]byte(out), []byte("\r ")) { // clear sequence contains "\r<spaces>\r"
		t.Fatalf("expected clear sequence, got %q", out)
	}
}

