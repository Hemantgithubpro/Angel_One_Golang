package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/e1/internal/auth"
	"example.com/e1/internal/config"
	"example.com/e1/internal/domain"
)

type Client struct {
	cfg        config.Config
	session    auth.Session
	logger     *log.Logger
	now        func() time.Time
	httpClient *http.Client
}

type quoteResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Fetched []struct {
			Exchange      string  `json:"exchange"`
			TradingSymbol string  `json:"tradingSymbol"`
			SymbolToken   string  `json:"symbolToken"`
			LTP           float64 `json:"ltp"`
		} `json:"fetched"`
	} `json:"data"`
}

func New(cfg config.Config, session auth.Session, logger *log.Logger) *Client {
	return &Client{
		cfg:        cfg,
		session:    session,
		logger:     logger,
		now:        time.Now,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) Run(ctx context.Context, out chan<- domain.Tick) error {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()

	if err := c.pollOnce(ctx, out); err != nil && c.logger != nil {
		c.logger.Printf("poller request failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := c.pollOnce(ctx, out); err != nil && c.logger != nil {
				c.logger.Printf("poller request failed: %v", err)
			}
		}
	}
}

func (c *Client) pollOnce(ctx context.Context, out chan<- domain.Tick) error {
	requestBody := map[string]any{
		"mode":           c.cfg.PollerMode,
		"exchangeTokens": buildExchangeTokens(c.cfg.PollerInstruments),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("marshal quote request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.QuoteURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create quote request: %w", err)
	}
	req.Header.Set("Authorization", c.session.JWTToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-UserType", "USER")
	req.Header.Set("X-SourceID", "WEB")
	req.Header.Set("X-ClientLocalIP", "127.0.0.1")
	req.Header.Set("X-ClientPublicIP", "127.0.0.1")
	req.Header.Set("X-MACAddress", "00:00:00:00:00:00")
	req.Header.Set("X-PrivateKey", c.session.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("quote request: %w", err)
	}
	defer resp.Body.Close()

	var result quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode quote response: %w", err)
	}
	if !result.Status {
		return fmt.Errorf("quote request failed: %s", result.Message)
	}

	now := c.now().UTC()
	for _, item := range result.Data.Fetched {
		out <- domain.Tick{
			Source:        domain.SourcePoller,
			Token:         item.SymbolToken,
			Exchange:      item.Exchange,
			TradingSymbol: item.TradingSymbol,
			EventTime:     now,
			ReceivedAt:    now,
			LTP:           item.LTP,
		}
	}
	return nil
}

func buildExchangeTokens(instruments []config.PollerInstrument) map[string][]string {
	exchangeTokens := make(map[string][]string)
	for _, instrument := range instruments {
		exchangeTokens[instrument.Exchange] = append(exchangeTokens[instrument.Exchange], instrument.SymbolToken)
	}
	return exchangeTokens
}
