package poller

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"example.com/e1/internal/auth"
	"example.com/e1/internal/config"
	"example.com/e1/internal/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestPollerMapsResponseToTicks(t *testing.T) {
	cfg := config.Config{
		PollerMode: "LTP",
		PollerInstruments: []config.PollerInstrument{
			{Exchange: "NSE", SymbolToken: "99926000"},
		},
		QuoteURL: "http://example.test/quote",
	}
	client := New(cfg, auth.Session{
		APIKey:   "key",
		JWTToken: "Bearer jwt",
	}, log.New(io.Discard, "", 0))
	client.now = func() time.Time { return time.Unix(1710000000, 0).UTC() }
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"status": true,
					"data": {
						"fetched": [
							{"exchange":"NSE","tradingSymbol":"NIFTY","symbolToken":"99926000","ltp":123.45}
						]
					}
				}`)),
			}, nil
		}),
	})

	out := make(chan domain.Tick, 1)
	if err := client.pollOnce(context.Background(), out); err != nil {
		t.Fatalf("pollOnce() error = %v", err)
	}

	tick := <-out
	if tick.Source != domain.SourcePoller || tick.Token != "99926000" || tick.LTP != 123.45 {
		t.Fatalf("unexpected tick: %+v", tick)
	}
}

func TestPollerHandlesEmptyFetchedData(t *testing.T) {
	cfg := config.Config{
		PollerMode:        "LTP",
		PollerInstruments: []config.PollerInstrument{{Exchange: "NSE", SymbolToken: "99926000"}},
		QuoteURL:          "http://example.test/quote",
	}
	client := New(cfg, auth.Session{APIKey: "key", JWTToken: "Bearer jwt"}, log.New(io.Discard, "", 0))
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status": true, "data": {"fetched": []}}`)),
			}, nil
		}),
	})

	out := make(chan domain.Tick, 1)
	if err := client.pollOnce(context.Background(), out); err != nil {
		t.Fatalf("pollOnce() error = %v", err)
	}
	select {
	case tick := <-out:
		t.Fatalf("did not expect tick: %+v", tick)
	default:
	}
}
