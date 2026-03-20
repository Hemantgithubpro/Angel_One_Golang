# Live Data Ingestor

This project runs a single Go service that logs into Angel One, reads live market data, and stores raw ticks in Postgres.

It supports two ingestion paths at the same time:

- `websocket`: Angel One Smart Stream live feed
- `poller`: repeated REST quote polling

Both sources write into the same append-only `live_ticks` table.

## What It Does

The service:

1. Loads configuration from environment variables or `.env`
2. Logs into Angel One using `API_KEY`, `CLIENT_ID`, `MPIN`, and `TOTP_SECRET`
3. Connects to Postgres
4. Creates the `live_ticks` table if it does not exist
5. Starts websocket ingestion, poller ingestion, or both
6. Buffers incoming ticks and writes them to Postgres in batches

The main entrypoint is [cmd/ingestor/main.go](/Users/hemant/Computing/algo_trading/angel_one/go_implementation/exp3/cmd/ingestor/main.go).

## Project Layout

- `cmd/ingestor`: service entrypoint
- `internal/config`: environment config loading and validation
- `internal/auth`: Angel One login and session creation
- `internal/ingest/websocket`: websocket ingestion and binary packet parsing
- `internal/ingest/poller`: REST polling ingestion
- `internal/service`: batching and shutdown-safe flushing
- `internal/storage/postgres`: Postgres schema setup and bulk inserts

Old prototype files still exist at the repo root, but they are excluded from the default build with `//go:build ignore`.

## Requirements

- Go installed locally
- Postgres available and reachable from `DB_URL`
- Angel One credentials

## Quick Start

1. Create a `.env` file from `.env.example`
2. Fill in your real credentials and database URL
3. Run:

```bash
go run ./cmd/ingestor
```

Or with Docker Compose:

```bash
docker compose up
```

## Environment Variables

Required:

- `DB_URL`: Postgres connection string
- `API_KEY`: Angel One API key
- `CLIENT_ID`: Angel One client code
- `MPIN`: Angel One MPIN/password
- `TOTP_SECRET`: TOTP secret for login

Ingestion toggles:

- `ENABLE_WEBSOCKET`: `true` or `false`, default `true`
- `ENABLE_POLLER`: `true` or `false`, default `true`

Websocket settings:

- `WEBSOCKET_MODE`: default `2`
- `WEBSOCKET_TOKENS`: JSON array of websocket subscriptions
- `WEBSOCKET_PING_PERIOD`: default `30s`
- `WEBSOCKET_URL`: defaults to Angel One Smart Stream URL

Poller settings:

- `POLLER_MODE`: default `LTP`
- `POLLER_INSTRUMENTS`: JSON array of instruments to poll
- `POLL_INTERVAL`: default `1s`
- `QUOTE_URL`: defaults to Angel One quote endpoint

Batching/runtime:

- `BATCH_SIZE`: default `500`
- `FLUSH_INTERVAL`: default `5s`
- `QUEUE_SIZE`: default `2048`
- `LOGIN_URL`: override Angel One login endpoint if needed

## Token Configuration Format

`WEBSOCKET_TOKENS` must be a JSON array like:

```json
[
  {
    "exchange_type": 1,
    "tokens": ["99926000", "2885"]
  },
  {
    "exchange_type": 3,
    "tokens": ["99919000"]
  }
]
```

`POLLER_INSTRUMENTS` must be a JSON array like:

```json
[
  {
    "exchange": "NSE",
    "symbol_token": "99926000"
  },
  {
    "exchange": "NSE",
    "symbol_token": "2885"
  }
]
```

## Database Schema

The service creates a `live_ticks` table automatically. Important columns:

- `source`
- `token`
- `exchange`
- `exchange_type`
- `trading_symbol`
- `event_time`
- `received_at`
- `ltp`
- `volume`
- `open_price`
- `high_price`
- `low_price`
- `close_price`
- `total_buy_qty`
- `total_sell_qty`
- `avg_traded_price`
- `upper_circuit`
- `lower_circuit`
- `high_52_week`
- `low_52_week`

Schema creation and inserts are handled in [store.go](/Users/hemant/Computing/algo_trading/angel_one/go_implementation/exp3/internal/storage/postgres/store.go).

## Common Usage Patterns

Run websocket only:

```env
ENABLE_WEBSOCKET=true
ENABLE_POLLER=false
```

Run poller only:

```env
ENABLE_WEBSOCKET=false
ENABLE_POLLER=true
```

Run both:

```env
ENABLE_WEBSOCKET=true
ENABLE_POLLER=true
```

## Verify It Is Working

Start the service, then check Postgres:

```sql
select source, token, event_time, ltp
from live_ticks
order by id desc
limit 20;
```

If rows are appearing, ingestion is working.

## Run Tests

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```

The local `GOCACHE` override is useful if the default Go cache path is restricted in your environment.

## Troubleshooting

`load config: ... is required`

- A required env var is missing

`login failed: ...`

- Angel One credentials, TOTP secret, or API key are wrong

No rows in `live_ticks`

- Check token JSON formatting first
- Check that the selected exchanges and symbol tokens are valid
- Check Postgres connectivity from `DB_URL`
- If websocket is enabled, confirm your feed tokens and subscription mode are accepted

Only poller rows or only websocket rows appear

- That usually means one ingestion path is working and the other is failing; inspect service logs
