package postgres

import (
	"context"
	"fmt"

	"example.com/e1/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(dbURL string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) InitSchema(ctx context.Context) error {
	const query = `
	CREATE TABLE IF NOT EXISTS live_ticks (
		id BIGSERIAL PRIMARY KEY,
		source TEXT NOT NULL,
		token TEXT NOT NULL,
		exchange TEXT NOT NULL DEFAULT '',
		exchange_type INT NOT NULL DEFAULT 0,
		trading_symbol TEXT NOT NULL DEFAULT '',
		event_time TIMESTAMPTZ NOT NULL,
		received_at TIMESTAMPTZ NOT NULL,
		ltp DOUBLE PRECISION NOT NULL,
		last_traded_qty BIGINT NOT NULL DEFAULT 0,
		volume BIGINT NOT NULL DEFAULT 0,
		open_price DOUBLE PRECISION NOT NULL DEFAULT 0,
		high_price DOUBLE PRECISION NOT NULL DEFAULT 0,
		low_price DOUBLE PRECISION NOT NULL DEFAULT 0,
		close_price DOUBLE PRECISION NOT NULL DEFAULT 0,
		total_buy_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
		total_sell_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
		avg_traded_price DOUBLE PRECISION NOT NULL DEFAULT 0,
		upper_circuit DOUBLE PRECISION NOT NULL DEFAULT 0,
		lower_circuit DOUBLE PRECISION NOT NULL DEFAULT 0,
		high_52_week DOUBLE PRECISION NOT NULL DEFAULT 0,
		low_52_week DOUBLE PRECISION NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_live_ticks_event_time ON live_ticks (event_time);
	CREATE INDEX IF NOT EXISTS idx_live_ticks_token_event_time ON live_ticks (token, event_time DESC);
	`
	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Store) WriteBatch(ctx context.Context, ticks []domain.Tick) error {
	if len(ticks) == 0 {
		return nil
	}
	rows := make([][]any, 0, len(ticks))
	for _, tick := range ticks {
		rows = append(rows, []any{
			string(tick.Source),
			tick.Token,
			tick.Exchange,
			tick.ExchangeType,
			tick.TradingSymbol,
			tick.EventTime,
			tick.ReceivedAt,
			tick.LTP,
			tick.LastTradedQty,
			tick.Volume,
			tick.OpenPrice,
			tick.HighPrice,
			tick.LowPrice,
			tick.ClosePrice,
			tick.TotalBuyQty,
			tick.TotalSellQty,
			tick.AvgTradedPrice,
			tick.UpperCircuit,
			tick.LowerCircuit,
			tick.High52Week,
			tick.Low52Week,
		})
	}

	_, err := s.pool.CopyFrom(
		ctx,
		pgx.Identifier{"live_ticks"},
		[]string{
			"source",
			"token",
			"exchange",
			"exchange_type",
			"trading_symbol",
			"event_time",
			"received_at",
			"ltp",
			"last_traded_qty",
			"volume",
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"total_buy_qty",
			"total_sell_qty",
			"avg_traded_price",
			"upper_circuit",
			"lower_circuit",
			"high_52_week",
			"low_52_week",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (s *Store) Close() {
	s.pool.Close()
}
