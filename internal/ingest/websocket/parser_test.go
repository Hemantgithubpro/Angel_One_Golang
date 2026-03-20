package websocket

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"example.com/e1/internal/domain"
)

func TestParseBinaryTickLTP(t *testing.T) {
	payload := make([]byte, 51)
	payload[0] = 1
	payload[1] = 1
	copy(payload[2:], []byte("99926000"))
	binary.LittleEndian.PutUint64(payload[35:43], uint64(1710000000000))
	binary.LittleEndian.PutUint64(payload[43:51], uint64(245678))

	now := func() time.Time { return time.Unix(1710000001, 0).UTC() }
	tick, err := ParseBinaryTick(payload, now)
	if err != nil {
		t.Fatalf("ParseBinaryTick() error = %v", err)
	}
	if tick.Source != domain.SourceWebsocket {
		t.Fatalf("unexpected source: %s", tick.Source)
	}
	if tick.Token != "99926000" {
		t.Fatalf("unexpected token: %q", tick.Token)
	}
	if tick.LTP != 2456.78 {
		t.Fatalf("unexpected ltp: %f", tick.LTP)
	}
	if !tick.ReceivedAt.Equal(now()) {
		t.Fatalf("unexpected received time: %s", tick.ReceivedAt)
	}
}

func TestParseBinaryTickQuote(t *testing.T) {
	payload := make([]byte, 123)
	payload[0] = 2
	payload[1] = 1
	copy(payload[2:], []byte("2885"))
	binary.LittleEndian.PutUint64(payload[35:43], uint64(1710000000000))
	binary.LittleEndian.PutUint64(payload[43:51], uint64(15025))
	binary.LittleEndian.PutUint64(payload[51:59], uint64(42))
	binary.LittleEndian.PutUint64(payload[59:67], uint64(15000))
	binary.LittleEndian.PutUint64(payload[67:75], uint64(900))
	binary.LittleEndian.PutUint64(payload[75:83], math.Float64bits(100))
	binary.LittleEndian.PutUint64(payload[83:91], math.Float64bits(200))
	binary.LittleEndian.PutUint64(payload[91:99], uint64(14900))
	binary.LittleEndian.PutUint64(payload[99:107], uint64(15100))
	binary.LittleEndian.PutUint64(payload[107:115], uint64(14800))
	binary.LittleEndian.PutUint64(payload[115:123], uint64(14950))

	tick, err := ParseBinaryTick(payload, time.Now)
	if err != nil {
		t.Fatalf("ParseBinaryTick() error = %v", err)
	}
	if tick.LastTradedQty != 42 || tick.Volume != 900 {
		t.Fatalf("unexpected qty fields: %+v", tick)
	}
	if tick.TotalBuyQty != 100 || tick.TotalSellQty != 200 {
		t.Fatalf("unexpected depth totals: %+v", tick)
	}
}

func TestParseBinaryTickSnapQuote(t *testing.T) {
	payload := make([]byte, 379)
	payload[0] = 3
	payload[1] = 1
	copy(payload[2:], []byte("2885"))
	binary.LittleEndian.PutUint64(payload[35:43], uint64(1710000000000))
	binary.LittleEndian.PutUint64(payload[43:51], uint64(15025))
	binary.LittleEndian.PutUint64(payload[347:355], uint64(15500))
	binary.LittleEndian.PutUint64(payload[355:363], uint64(14000))
	binary.LittleEndian.PutUint64(payload[363:371], uint64(16000))
	binary.LittleEndian.PutUint64(payload[371:379], uint64(13000))

	tick, err := ParseBinaryTick(payload, time.Now)
	if err != nil {
		t.Fatalf("ParseBinaryTick() error = %v", err)
	}
	if tick.UpperCircuit != 155 || tick.Low52Week != 130 {
		t.Fatalf("unexpected snap fields: %+v", tick)
	}
}

func TestParseBinaryTickRejectsInvalidPayload(t *testing.T) {
	if _, err := ParseBinaryTick([]byte{2}, time.Now); err == nil {
		t.Fatal("expected invalid packet size error")
	}
	if _, err := ParseBinaryTick([]byte{9, 1, 2}, time.Now); err == nil {
		t.Fatal("expected unknown mode error")
	}
}
