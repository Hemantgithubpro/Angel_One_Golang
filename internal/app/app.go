package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"example.com/e1/internal/domain"
	"example.com/e1/internal/ingest"
	"example.com/e1/internal/service"
)

type Ingestor = ingest.Ingestor

type Options struct {
	Logger        *log.Logger
	Writer        service.BatchWriter
	Ingestors     []Ingestor
	BatchSize     int
	FlushInterval time.Duration
	QueueSize     int
}

type App struct {
	logger        *log.Logger
	writer        service.BatchWriter
	ingestors     []Ingestor
	batchSize     int
	flushInterval time.Duration
	queueSize     int
}

func New(opts Options) *App {
	return &App{
		logger:        opts.Logger,
		writer:        opts.Writer,
		ingestors:     opts.Ingestors,
		batchSize:     opts.BatchSize,
		flushInterval: opts.FlushInterval,
		queueSize:     opts.QueueSize,
	}
}

func (a *App) Run(ctx context.Context) error {
	if a.writer == nil {
		return fmt.Errorf("writer is required")
	}
	if len(a.ingestors) == 0 {
		return fmt.Errorf("at least one ingestor is required")
	}

	ticks := make(chan domain.Tick, a.queueSize)
	batcher := service.NewBatcher(a.writer, a.batchSize, a.flushInterval, a.logger)

	batcherDone := make(chan error, 1)
	go func() {
		batcherDone <- batcher.Run(ctx, ticks)
	}()

	var wg sync.WaitGroup
	for _, ingestor := range a.ingestors {
		ingestor := ingestor
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ingestor.Run(ctx, ticks); err != nil && a.logger != nil {
				a.logger.Printf("ingestor stopped with error: %v", err)
			}
		}()
	}

	select {
	case err := <-batcherDone:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
	}

	wg.Wait()
	close(ticks)

	if err := <-batcherDone; err != nil {
		return err
	}
	return nil
}
