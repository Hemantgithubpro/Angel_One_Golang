package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/e1/internal/domain"
)

type BatchWriter interface {
	WriteBatch(ctx context.Context, ticks []domain.Tick) error
}

type Batcher struct {
	writer        BatchWriter
	batchSize     int
	flushInterval time.Duration
	logger        *log.Logger
}

func NewBatcher(writer BatchWriter, batchSize int, flushInterval time.Duration, logger *log.Logger) *Batcher {
	return &Batcher{
		writer:        writer,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		logger:        logger,
	}
}

func (b *Batcher) Run(ctx context.Context, in <-chan domain.Tick) error {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([]domain.Tick, 0, b.batchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		toWrite := append([]domain.Tick(nil), batch...)
		batch = batch[:0]
		if err := b.writer.WriteBatch(ctx, toWrite); err != nil {
			return fmt.Errorf("write batch: %w", err)
		}
		return nil
	}

	for {
		select {
		case tick, ok := <-in:
			if !ok {
				return flush()
			}
			batch = append(batch, tick)
			if len(batch) >= b.batchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if err := flush(); err != nil {
				return err
			}
		case <-ctx.Done():
			for {
				select {
				case tick, ok := <-in:
					if !ok {
						return flush()
					}
					batch = append(batch, tick)
				default:
					return flush()
				}
			}
		}
	}
}
