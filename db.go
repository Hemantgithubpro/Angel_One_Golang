//go:build ignore

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewDatabase() (*Database, error) {
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DB_URL environment variable is required")
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

func (db *Database) InsertTick(ctx context.Context, tick MarketTick) error {
	query := `
	INSERT INTO market_ticks (token, exchange_type, timestamp, price, volume, open_price, high_price, low_price, close_price, total_buy_qty, total_sell_qty, avg_traded_price)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := db.Pool.Exec(ctx, query,
		tick.Token,
		tick.ExchangeType,
		tick.Timestamp,
		tick.LTP,
		tick.Volume,
		tick.OpenPrice,
		tick.HighPrice,
		tick.LowPrice,
		tick.ClosePrice,
		tick.TotalBuyQty,
		tick.TotalSellQty,
		tick.AvgTradedPrice,
	)
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

func trydb() {
	// make databse connection and initialize schema
	db, err := NewDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.InitSchema(ctx); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	sampleTick := MarketTick{
		Token:          "99919000",
		ExchangeType:   3,
		Timestamp:      time.Now(),
		LTP:            100.5,
		Volume:         1000,
		OpenPrice:      99.0,
		HighPrice:      101.0,
		LowPrice:       98.5,
		ClosePrice:     100.5,
		TotalBuyQty:    5000,
		TotalSellQty:   4500,
		AvgTradedPrice: 100.0,
	}
	if err := db.InsertTick(ctx, sampleTick); err != nil {
		log.Fatal("Failed to insert sample tick:", err)
	}

}

func exportDatatoCSV(db *Database) error {
	ctx := context.Background()
	date := time.Now().Format("2006-01-02")
	datetimestart:= date + " 03:44:59"
	datetimeend:= date + ""

	rows, err := db.Pool.Query(ctx, "SELECT token, exchange_type, timestamp, price, volume, open_price, high_price, low_price, close_price, total_buy_qty, total_sell_qty, avg_traded_price FROM market_ticks")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	filename := fmt.Sprintf("market_ticks_%s.csv", date)
	// Create CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV headers
	headers := []string{"id", "token", "exchange_type", "timestamp", "price", "volume", "open_price", "high_price", "low_price", "close_price", "total_buy_qty", "total_sell_qty", "avg_traded_price", "created_at"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write rows
	for rows.Next() {
		var id int64
		var token string
		var exchangeType int
		var timestamp time.Time
		var price float64
		var volume int64
		var openPrice float64
		var highPrice float64
		var lowPrice float64
		var closePrice float64
		var totalBuyQty int64
		var totalSellQty int64
		var avgTradedPrice float64
		var createdAt time.Time

		if err := rows.Scan(&id, &token, &exchangeType, &timestamp, &price, &volume, &openPrice, &highPrice, &lowPrice, &closePrice, &totalBuyQty, &totalSellQty, &avgTradedPrice, &createdAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		row := []string{fmt.Sprintf("%d", id), token, fmt.Sprintf("%d", exchangeType), timestamp.Format(time.RFC3339), fmt.Sprintf("%.2f", price), fmt.Sprintf("%d", volume), fmt.Sprintf("%.2f", openPrice), fmt.Sprintf("%.2f", highPrice), fmt.Sprintf("%.2f", lowPrice), fmt.Sprintf("%.2f", closePrice), fmt.Sprintf("%d", totalBuyQty), fmt.Sprintf("%d", totalSellQty), fmt.Sprintf("%.2f", avgTradedPrice), createdAt.Format(time.RFC3339)}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	log.Printf("Successfully exported data to CSV file: %s", filename)
	return nil
}
