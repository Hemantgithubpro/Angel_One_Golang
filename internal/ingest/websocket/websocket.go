package websocket

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"example.com/e1/internal/auth"
	"example.com/e1/internal/config"
	"example.com/e1/internal/domain"
	"github.com/gorilla/websocket"
)

type Client struct {
	cfg     config.Config
	session auth.Session
	logger  *log.Logger
	now     func() time.Time
	dialer  *websocket.Dialer
}

type streamRequest struct {
	CorrelationID string `json:"correlationID"`
	Action        int    `json:"action"`
	Params        struct {
		Mode      int `json:"mode"`
		TokenList []struct {
			ExchangeType int      `json:"exchangeType"`
			Tokens       []string `json:"tokens"`
		} `json:"tokenList"`
	} `json:"params"`
}

func New(cfg config.Config, session auth.Session, logger *log.Logger) *Client {
	return &Client{
		cfg:     cfg,
		session: session,
		logger:  logger,
		now:     time.Now,
		dialer:  websocket.DefaultDialer,
	}
}

func (c *Client) Run(ctx context.Context, out chan<- domain.Tick) error {
	headers := http.Header{}
	headers.Set("Authorization", c.session.JWTToken)
	headers.Set("x-api-key", c.session.APIKey)
	headers.Set("x-client-code", c.session.ClientID)
	headers.Set("x-feed-token", c.session.FeedToken)

	conn, _, err := c.dialer.DialContext(ctx, c.cfg.WebsocketURL, headers)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(buildRequest(c.cfg)); err != nil {
		return fmt.Errorf("subscribe websocket: %w", err)
	}

	pingTicker := time.NewTicker(c.cfg.WebsocketPingPeriod)
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil && c.logger != nil {
					c.logger.Printf("websocket ping failed: %v", err)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return nil
		default:
		}

		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("read websocket message: %w", err)
		}
		if messageType != websocket.BinaryMessage {
			continue
		}

		tick, err := ParseBinaryTick(payload, c.now)
		if err != nil {
			if c.logger != nil {
				c.logger.Printf("parse websocket payload: %v", err)
			}
			continue
		}
		if tick == nil {
			continue
		}
		out <- *tick
	}
}

func buildRequest(cfg config.Config) streamRequest {
	req := streamRequest{
		CorrelationID: "live-ingestor",
		Action:        1,
	}
	req.Params.Mode = cfg.WebsocketMode
	for _, sub := range cfg.WebsocketTokens {
		req.Params.TokenList = append(req.Params.TokenList, struct {
			ExchangeType int      `json:"exchangeType"`
			Tokens       []string `json:"tokens"`
		}{
			ExchangeType: sub.ExchangeType,
			Tokens:       append([]string(nil), sub.Tokens...),
		})
	}
	return req
}

func ParseBinaryTick(data []byte, now func() time.Time) (*domain.Tick, error) {
	if len(data) == 0 {
		return nil, nil
	}
	if now == nil {
		now = time.Now
	}

	switch data[0] {
	case 1:
		if len(data) != 51 {
			return nil, fmt.Errorf("invalid LTP packet size: %d", len(data))
		}
		return parseLTPPacket(data, now), nil
	case 2:
		if len(data) != 123 {
			return nil, fmt.Errorf("invalid quote packet size: %d", len(data))
		}
		return parseQuotePacket(data, now), nil
	case 3:
		if len(data) != 379 {
			return nil, fmt.Errorf("invalid snap quote packet size: %d", len(data))
		}
		return parseSnapQuotePacket(data, now), nil
	default:
		return nil, fmt.Errorf("unknown subscription mode: %d", data[0])
	}
}

func parseLTPPacket(data []byte, now func() time.Time) *domain.Tick {
	exchangeType := int(data[1])
	divisor := priceDivisor(exchangeType)
	return &domain.Tick{
		Source:       domain.SourceWebsocket,
		Token:        string(bytes.Trim(data[2:27], "\x00")),
		ExchangeType: exchangeType,
		EventTime:    time.UnixMilli(int64(binary.LittleEndian.Uint64(data[35:43]))),
		ReceivedAt:   now().UTC(),
		LTP:          float64(int64(binary.LittleEndian.Uint64(data[43:51]))) / divisor,
	}
}

func parseQuotePacket(data []byte, now func() time.Time) *domain.Tick {
	exchangeType := int(data[1])
	divisor := priceDivisor(exchangeType)
	return &domain.Tick{
		Source:         domain.SourceWebsocket,
		Token:          string(bytes.Trim(data[2:27], "\x00")),
		ExchangeType:   exchangeType,
		EventTime:      time.UnixMilli(int64(binary.LittleEndian.Uint64(data[35:43]))),
		ReceivedAt:     now().UTC(),
		LTP:            float64(int64(binary.LittleEndian.Uint64(data[43:51]))) / divisor,
		LastTradedQty:  int64(binary.LittleEndian.Uint64(data[51:59])),
		AvgTradedPrice: float64(int64(binary.LittleEndian.Uint64(data[59:67]))) / divisor,
		Volume:         int64(binary.LittleEndian.Uint64(data[67:75])),
		TotalBuyQty:    math.Float64frombits(binary.LittleEndian.Uint64(data[75:83])),
		TotalSellQty:   math.Float64frombits(binary.LittleEndian.Uint64(data[83:91])),
		OpenPrice:      float64(int64(binary.LittleEndian.Uint64(data[91:99]))) / divisor,
		HighPrice:      float64(int64(binary.LittleEndian.Uint64(data[99:107]))) / divisor,
		LowPrice:       float64(int64(binary.LittleEndian.Uint64(data[107:115]))) / divisor,
		ClosePrice:     float64(int64(binary.LittleEndian.Uint64(data[115:123]))) / divisor,
	}
}

func parseSnapQuotePacket(data []byte, now func() time.Time) *domain.Tick {
	tick := parseQuotePacket(data[:123], now)
	divisor := priceDivisor(int(data[1]))
	tick.UpperCircuit = float64(int64(binary.LittleEndian.Uint64(data[347:355]))) / divisor
	tick.LowerCircuit = float64(int64(binary.LittleEndian.Uint64(data[355:363]))) / divisor
	tick.High52Week = float64(int64(binary.LittleEndian.Uint64(data[363:371]))) / divisor
	tick.Low52Week = float64(int64(binary.LittleEndian.Uint64(data[371:379]))) / divisor
	return tick
}

func priceDivisor(exchangeType int) float64 {
	if exchangeType == 13 {
		return 10000000.0
	}
	return 100.0
}

func MarshalRequest(cfg config.Config) ([]byte, error) {
	return json.Marshal(buildRequest(cfg))
}
