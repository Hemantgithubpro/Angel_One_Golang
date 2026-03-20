package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaultsAndRequiredValues(t *testing.T) {
	t.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("API_KEY", "key")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("MPIN", "1234")
	t.Setenv("TOTP_SECRET", "secret")
	t.Setenv("WEBSOCKET_TOKENS", `[{"exchange_type":1,"tokens":["99926000"]}]`)
	t.Setenv("POLLER_INSTRUMENTS", `[{"exchange":"NSE","symbol_token":"99926000"}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.EnableWebsocket || !cfg.EnablePoller {
		t.Fatalf("expected both ingestors enabled by default")
	}
	if cfg.BatchSize != 500 {
		t.Fatalf("expected default batch size 500, got %d", cfg.BatchSize)
	}
	if cfg.FlushInterval != 5*time.Second {
		t.Fatalf("expected default flush interval 5s, got %s", cfg.FlushInterval)
	}
	if cfg.PollInterval != time.Second {
		t.Fatalf("expected default poll interval 1s, got %s", cfg.PollInterval)
	}
}

func TestLoadRequiresEnabledSourceConfig(t *testing.T) {
	t.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("API_KEY", "key")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("MPIN", "1234")
	t.Setenv("TOTP_SECRET", "secret")
	t.Setenv("ENABLE_WEBSOCKET", "false")
	t.Setenv("ENABLE_POLLER", "true")
	t.Setenv("POLLER_INSTRUMENTS", "")
	t.Setenv("WEBSOCKET_TOKENS", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load() to fail when poller config is missing")
	}
}

func TestLoadDisableModes(t *testing.T) {
	t.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("API_KEY", "key")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("MPIN", "1234")
	t.Setenv("TOTP_SECRET", "secret")
	t.Setenv("ENABLE_WEBSOCKET", "false")
	t.Setenv("ENABLE_POLLER", "false")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when both ingestors are disabled")
	}
}

func TestLoadParsesCustomDurations(t *testing.T) {
	t.Setenv("DB_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("API_KEY", "key")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("MPIN", "1234")
	t.Setenv("TOTP_SECRET", "secret")
	t.Setenv("WEBSOCKET_TOKENS", `[{"exchange_type":1,"tokens":["99926000"]}]`)
	t.Setenv("POLLER_INSTRUMENTS", `[{"exchange":"NSE","symbol_token":"99926000"}]`)
	t.Setenv("POLL_INTERVAL", "2s")
	t.Setenv("FLUSH_INTERVAL", "9s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.PollInterval != 2*time.Second || cfg.FlushInterval != 9*time.Second {
		t.Fatalf("unexpected durations: poll=%s flush=%s", cfg.PollInterval, cfg.FlushInterval)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.Unsetenv("ENABLE_WEBSOCKET")
	_ = os.Unsetenv("ENABLE_POLLER")
	os.Exit(code)
}
