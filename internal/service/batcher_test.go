package service

import (
	"context"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"example.com/e1/internal/domain"
)

type recordingWriter struct {
	mu      sync.Mutex
	batches [][]domain.Tick
}

func (w *recordingWriter) WriteBatch(_ context.Context, ticks []domain.Tick) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.batches = append(w.batches, append([]domain.Tick(nil), ticks...))
	return nil
}

func TestBatcherFlushesOnBatchSize(t *testing.T) {
	writer := &recordingWriter{}
	batcher := NewBatcher(writer, 2, time.Hour, log.New(io.Discard, "", 0))
	input := make(chan domain.Tick, 2)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- batcher.Run(ctx, input) }()

	input <- domain.Tick{Token: "a"}
	input <- domain.Tick{Token: "b"}
	close(input)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(writer.batches) != 1 || len(writer.batches[0]) != 2 {
		t.Fatalf("unexpected batches: %+v", writer.batches)
	}
}

func TestBatcherFlushesOnTimer(t *testing.T) {
	writer := &recordingWriter{}
	batcher := NewBatcher(writer, 10, 20*time.Millisecond, log.New(io.Discard, "", 0))
	input := make(chan domain.Tick, 1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- batcher.Run(ctx, input) }()

	input <- domain.Tick{Token: "a"}
	time.Sleep(50 * time.Millisecond)
	close(input)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(writer.batches) == 0 {
		t.Fatal("expected timer flush")
	}
}

func TestBatcherFlushesOnShutdown(t *testing.T) {
	writer := &recordingWriter{}
	batcher := NewBatcher(writer, 10, time.Hour, log.New(io.Discard, "", 0))
	input := make(chan domain.Tick, 2)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- batcher.Run(ctx, input) }()

	input <- domain.Tick{Token: "a"}
	input <- domain.Tick{Token: "b"}
	cancel()
	close(input)

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(writer.batches) != 1 || len(writer.batches[0]) != 2 {
		t.Fatalf("expected shutdown flush, got %+v", writer.batches)
	}
}
