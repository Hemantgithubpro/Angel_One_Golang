package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/e1/internal/app"
	"example.com/e1/internal/auth"
	"example.com/e1/internal/config"
	"example.com/e1/internal/ingest/poller"
	ws "example.com/e1/internal/ingest/websocket"
	"example.com/e1/internal/storage/postgres"
)

func main() {
	logger := log.New(os.Stdout, "ingestor ", log.LstdFlags|log.Lmicroseconds|log.LUTC)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	sessionClient := auth.NewClient(cfg.APIKey, cfg.ClientID, cfg.MPIN, cfg.TOTPSecret)
	sessionClient.SetLoginURL(cfg.LoginURL)
	session, err := sessionClient.Login(context.Background())
	if err != nil {
		logger.Fatalf("login: %v", err)
	}

	store, err := postgres.NewStore(cfg.DBURL)
	if err != nil {
		logger.Fatalf("create store: %v", err)
	}
	defer store.Close()

	if err := store.InitSchema(context.Background()); err != nil {
		logger.Fatalf("init schema: %v", err)
	}

	ingestors := make([]app.Ingestor, 0, 2)
	if cfg.EnableWebsocket {
		ingestors = append(ingestors, ws.New(cfg, session, logger))
	}
	if cfg.EnablePoller {
		ingestors = append(ingestors, poller.New(cfg, session, logger))
	}

	service := app.New(app.Options{
		Logger:        logger,
		Writer:        store,
		Ingestors:     ingestors,
		BatchSize:     cfg.BatchSize,
		FlushInterval: cfg.FlushInterval,
		QueueSize:     cfg.QueueSize,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil {
		logger.Fatalf("run service: %v", err)
	}
}
