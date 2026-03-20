package ingest

import (
	"context"

	"example.com/e1/internal/domain"
)

type Ingestor interface {
	Run(ctx context.Context, out chan<- domain.Tick) error
}
