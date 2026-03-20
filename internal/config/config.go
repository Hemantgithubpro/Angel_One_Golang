package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type WebsocketSubscription struct {
	ExchangeType int      `json:"exchange_type"`
	Tokens       []string `json:"tokens"`
}

type PollerInstrument struct {
	Exchange    string `json:"exchange"`
	SymbolToken string `json:"symbol_token"`
}

type Config struct {
	DBURL               string
	APIKey              string
	ClientID            string
	MPIN                string
	TOTPSecret          string
	EnableWebsocket     bool
	EnablePoller        bool
	WebsocketMode       int
	WebsocketTokens     []WebsocketSubscription
	PollerMode          string
	PollerInstruments   []PollerInstrument
	PollInterval        time.Duration
	BatchSize           int
	FlushInterval       time.Duration
	QueueSize           int
	WebsocketURL        string
	QuoteURL            string
	LoginURL            string
	WebsocketPingPeriod time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		APIKey:              os.Getenv("API_KEY"),
		DBURL:               os.Getenv("DB_URL"),
		ClientID:            os.Getenv("CLIENT_ID"),
		MPIN:                os.Getenv("MPIN"),
		TOTPSecret:          os.Getenv("TOTP_SECRET"),
		EnableWebsocket:     getEnvBool("ENABLE_WEBSOCKET", true),
		EnablePoller:        getEnvBool("ENABLE_POLLER", true),
		WebsocketMode:       getEnvInt("WEBSOCKET_MODE", 2),
		PollerMode:          getEnvString("POLLER_MODE", "LTP"),
		PollInterval:        getEnvDuration("POLL_INTERVAL", time.Second),
		BatchSize:           getEnvInt("BATCH_SIZE", 500),
		FlushInterval:       getEnvDuration("FLUSH_INTERVAL", 5*time.Second),
		QueueSize:           getEnvInt("QUEUE_SIZE", 2048),
		WebsocketURL:        getEnvString("WEBSOCKET_URL", "wss://smartapisocket.angelone.in/smart-stream"),
		QuoteURL:            getEnvString("QUOTE_URL", "https://apiconnect.angelone.in/rest/secure/angelbroking/market/v1/quote/"),
		LoginURL:            getEnvString("LOGIN_URL", "https://apiconnect.angelone.in/rest/auth/angelbroking/user/v1/loginByPassword"),
		WebsocketPingPeriod: getEnvDuration("WEBSOCKET_PING_PERIOD", 30*time.Second),
	}

	if err := parseJSONEnv("WEBSOCKET_TOKENS", &cfg.WebsocketTokens); err != nil {
		return Config{}, err
	}
	if err := parseJSONEnv("POLLER_INSTRUMENTS", &cfg.PollerInstruments); err != nil {
		return Config{}, err
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	required := map[string]string{
		"DB_URL":      cfg.DBURL,
		"API_KEY":     cfg.APIKey,
		"CLIENT_ID":   cfg.ClientID,
		"MPIN":        cfg.MPIN,
		"TOTP_SECRET": cfg.TOTPSecret,
	}
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", key)
		}
	}
	if !cfg.EnableWebsocket && !cfg.EnablePoller {
		return fmt.Errorf("at least one ingestor must be enabled")
	}
	if cfg.EnableWebsocket && len(cfg.WebsocketTokens) == 0 {
		return fmt.Errorf("WEBSOCKET_TOKENS is required when websocket ingestion is enabled")
	}
	if cfg.EnablePoller && len(cfg.PollerInstruments) == 0 {
		return fmt.Errorf("POLLER_INSTRUMENTS is required when poller ingestion is enabled")
	}
	if cfg.BatchSize <= 0 {
		return fmt.Errorf("BATCH_SIZE must be > 0")
	}
	if cfg.QueueSize <= 0 {
		return fmt.Errorf("QUEUE_SIZE must be > 0")
	}
	if cfg.FlushInterval <= 0 {
		return fmt.Errorf("FLUSH_INTERVAL must be > 0")
	}
	if cfg.PollInterval <= 0 {
		return fmt.Errorf("POLL_INTERVAL must be > 0")
	}
	if cfg.WebsocketPingPeriod <= 0 {
		return fmt.Errorf("WEBSOCKET_PING_PERIOD must be > 0")
	}
	return nil
}

func parseJSONEnv(name string, target any) error {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

func getEnvString(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvBool(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}
