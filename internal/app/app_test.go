package app

import (
	"context"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"example.com/e1/internal/domain"
)

type fakeIngestor struct {
	run func(ctx context.Context, out chan<- domain.Tick) error
}

func (f fakeIngestor) Run(ctx context.Context, out chan<- domain.Tick) error {
	return f.run(ctx, out)
}

type fakeWriter struct {
	mu    sync.Mutex
	ticks []domain.Tick
}

func (w *fakeWriter) WriteBatch(_ context.Context, ticks []domain.Tick) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ticks = append(w.ticks, ticks...)
	return nil
}

func TestAppKeepsRunningWhenOneIngestorFails(t *testing.T) {
	writer := &fakeWriter{}
	slowStarted := make(chan struct{})

	app := New(Options{
		Logger: log.New(io.Discard, "", 0),
		Writer: writer,
		Ingestors: []Ingestor{
			fakeIngestor{run: func(ctx context.Context, out chan<- domain.Tick) error {
				return context.Canceled
			}},
			fakeIngestor{run: func(ctx context.Context, out chan<- domain.Tick) error {
				close(slowStarted)
				<-ctx.Done()
				return nil
			}},
		},
		BatchSize:     1,
		FlushInterval: time.Second,
		QueueSize:     4,
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	select {
	case <-slowStarted:
	case <-time.After(time.Second):
		t.Fatal("second ingestor did not start")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestAppFlushesTicksOnShutdown(t *testing.T) {
	writer := &fakeWriter{}
	app := New(Options{
		Logger: log.New(io.Discard, "", 0),
		Writer: writer,
		Ingestors: []Ingestor{
			fakeIngestor{run: func(ctx context.Context, out chan<- domain.Tick) error {
				out <- domain.Tick{Token: "a"}
				<-ctx.Done()
				return nil
			}},
		},
		BatchSize:     10,
		FlushInterval: time.Hour,
		QueueSize:     4,
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(writer.ticks) != 1 {
		t.Fatalf("expected flushed tick on shutdown, got %+v", writer.ticks)
	}
}
