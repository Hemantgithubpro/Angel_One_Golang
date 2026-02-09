//go:build websocket_example

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewDatabase() (*Database, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &Database{Pool: pool}, nil
}

func (db *Database) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS market_ticks (
		id BIGSERIAL PRIMARY KEY,
		token TEXT NOT NULL,
		exchange_type INT,
		timestamp TIMESTAMPTZ,
		price DOUBLE PRECISION,
		volume BIGINT,
		open_price DOUBLE PRECISION,
		high_price DOUBLE PRECISION,
		low_price DOUBLE PRECISION,
		close_price DOUBLE PRECISION,
		total_buy_qty DOUBLE PRECISION,
		total_sell_qty DOUBLE PRECISION,
		avg_traded_price DOUBLE PRECISION,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);
	`
	_, err := db.Pool.Exec(ctx, query)
	return err
}

func (db *Database) BulkInsert(ctx context.Context, ticks []MarketTick) error {
	if len(ticks) == 0 {
		return nil
	}

	rows := [][]interface{}{}
	for _, t := range ticks {
		// Only mapping common fields for now, extend as needed
		rows = append(rows, []interface{}{
			t.Token, t.ExchangeType, t.Timestamp, t.LTP,
			t.Volume, t.OpenPrice, t.HighPrice, t.LowPrice, t.ClosePrice,
			t.TotalBuyQty, t.TotalSellQty, t.AvgTradedPrice,
		})
	}

	copyCount, err := db.Pool.CopyFrom(
		ctx,
		pgx.Identifier{"market_ticks"},
		[]string{
			"token", "exchange_type", "timestamp", "price",
			"volume", "open_price", "high_price", "low_price", "close_price",
			"total_buy_qty", "total_sell_qty", "avg_traded_price",
		},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return err
	}
	log.Printf("Batched %d records to DB", copyCount)
	return nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
