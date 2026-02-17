// Simple Websocket connection and subscription example. make http request first to get tokens

package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Config represents target URLs
const (
	WebsocketURL = "wss://smartapisocket.angelone.in/smart-stream"
)

// Angel One Stream Request Structs
type TokenInfo struct {
	ExchangeType int      `json:"exchangeType"`
	Tokens       []string `json:"tokens"`
}

type StreamParams struct {
	Mode      int         `json:"mode"`
	TokenList []TokenInfo `json:"tokenList"`
}

type StreamRequest struct {
	CorrelationID string       `json:"correlationID"`
	Action        int          `json:"action"`
	Params        StreamParams `json:"params"`
}

func websocketConnection1(jwt_token string, api_key string, client_id string, feed_token string, mode int, token TokenInfo, buffer *TickBuffer) {
	// --- STEP 1: Websocket Connection ---
	log.Println("Step 1: Connecting to WebSocket...")

	// Set up the interrupt channel to handle Ctrl+C gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create request headers containing the token
	headers := http.Header{}
	headers.Add("Authorization", jwt_token)
	headers.Add("x-api-key", api_key)
	headers.Add("x-client-code", client_id)
	headers.Add("x-feed-token", feed_token)

	// Dial the connection
	conn, resp, err := websocket.DefaultDialer.Dial(WebsocketURL, headers)
	if err != nil {
		if resp != nil {
			log.Printf("Handshake status: %d", resp.StatusCode)
		}
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket.")

	// --- STEP 2: Subscribe ---
	// Create the subscription request object
	req := StreamRequest{
		CorrelationID: "abcde12345",
		Action:        1, // Subscribe
		Params: StreamParams{
			Mode: mode, // 1 LTP Mode
			// Mode: 2, // Quote Mode (contains LTP + ohlc + volume + buy/sell qty + atp (average traded price))
			// Mode: 3, // Snap Quote Mode (contains everything in Quote + upper/lower circuit limits + 52 week high/low)
			// TokenList: []TokenInfo{
			// 	{
			// 		// ExchangeType: 1,                    // NSE
			// 		// Tokens:       []string{"99926000"}, // Nifty 50
			// 		// ExchangeType: 2,                    // NFO
			// 		// Tokens:       []string{"48236"}, // Nifty 50 Future
			// 		// ExchangeType: 3,                    // BSE
			// 		// Tokens:       []string{"99919000"}, // sensex
			// 	},
			// },
			TokenList: []TokenInfo{token},
		},
	}

	// Send the request
	log.Println("Sending subscription request...")
	err = conn.WriteJSON(req)
	if err != nil {
		log.Printf("Subscription failed: %v", err)
		return
	}
	log.Println("Subscribed to tokens.")

	// --- STEP 3: Read Loop (Background) ---
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			switch messageType {
			case websocket.TextMessage:
				log.Printf("Received Text: %s", string(message))
			case websocket.BinaryMessage:
				parseBinaryResponse(message, buffer)
			}
		}
	}()

	// --- STEP 4: Heartbeat Loop ---
	// Send 'ping' every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// --- STEP 5: Main Loop (Keep Alive / Shutdown) ---
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Send heartbeat
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Println("Heartbeat error:", err)
				return
			}
			log.Println("Sent Heartbeat: ping")
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")

			// Cleanly close the connection by sending a Close message
			err := conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				log.Println("Write close error:", err)
				return
			}

			// Wait a brief moment for the server to acknowledge the close
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func parseBinaryResponse(data []byte, buffer *TickBuffer) {
	if len(data) == 0 {
		return
	}

	mode := data[0]
	// exchangeType := data[1]
	// To distinguish currency for divisor, we might need exchangeType.
	// 1=nse_cm, 2=nse_fo, 3=bse_cm, 4=bse_fo, 5=mcx_fo, 7=ncx_fo, 13=cde_fo
	// For simplicity using 100.0 divisor. Real implementation should check ExchangeType 13.

	var tick *MarketTick
	var err error

	switch mode {
	case 1: // LTP Mode
		if len(data) != 51 {
			log.Printf("Invalid LTP packet size: %d", len(data))
			return
		}
		tick, err = parseLTPPacket(data)
	case 2: // Quote Mode
		if len(data) != 123 {
			log.Printf("Invalid Quote packet size: %d", len(data))
			return
		}
		tick, err = parseQuotePacket(data)
	case 3: // Snap Quote Mode
		if len(data) != 379 {
			log.Printf("Invalid SnapQuote packet size: %d", len(data))
			return
		}
		tick, err = parseSnapQuotePacket(data)
	default:
		log.Printf("Unknown Subscription Mode: %d", mode)
		return
	}

	if err != nil {
		log.Printf("Parse error: %v", err)
	} else if tick != nil {
		buffer.Add(*tick)
	}
}

func parseLTPPacket(data []byte) (*MarketTick, error) {
	exchangeType := data[1]
	token := string(bytes.Trim(data[2:27], "\x00"))
	// seqNum := int64(binary.LittleEndian.Uint64(data[27:35]))
	exchangeTime := int64(binary.LittleEndian.Uint64(data[35:43]))
	ltp := int64(binary.LittleEndian.Uint64(data[43:51]))

	divisor := 100.0
	if exchangeType == 13 {
		divisor = 10000000.0
	}
	realLTP := float64(ltp) / divisor
	tm := time.UnixMilli(exchangeTime)

	return &MarketTick{
		Token:        token,
		ExchangeType: int(exchangeType),
		Timestamp:    tm,
		LTP:          realLTP,
	}, nil
}

func parseQuotePacket(data []byte) (*MarketTick, error) {
	// Re-use headers from LTP part
	exchangeType := data[1]
	token := string(bytes.Trim(data[2:27], "\x00"))
	// seqNum := int64(binary.LittleEndian.Uint64(data[27:35]))
	exchangeTime := int64(binary.LittleEndian.Uint64(data[35:43]))
	ltp := int64(binary.LittleEndian.Uint64(data[43:51]))

	// Additional Quote Fields
	// lastTradedQty := int64(binary.LittleEndian.Uint64(data[51:59]))
	avgTradedPrice := int64(binary.LittleEndian.Uint64(data[59:67]))
	volTraded := int64(binary.LittleEndian.Uint64(data[67:75]))
	totalBuyQty := mathFloat64frombits(binary.LittleEndian.Uint64(data[75:83]))
	totalSellQty := mathFloat64frombits(binary.LittleEndian.Uint64(data[83:91]))
	openPrice := int64(binary.LittleEndian.Uint64(data[91:99]))
	highPrice := int64(binary.LittleEndian.Uint64(data[99:107]))
	lowPrice := int64(binary.LittleEndian.Uint64(data[107:115]))
	closePrice := int64(binary.LittleEndian.Uint64(data[115:123]))

	divisor := 100.0
	if exchangeType == 13 {
		divisor = 10000000.0
	}

	tm := time.UnixMilli(exchangeTime)

	return &MarketTick{
		Token:          token,
		ExchangeType:   int(exchangeType),
		Timestamp:      tm,
		LTP:            float64(ltp) / divisor,
		Volume:         volTraded,
		OpenPrice:      float64(openPrice) / divisor,
		HighPrice:      float64(highPrice) / divisor,
		LowPrice:       float64(lowPrice) / divisor,
		ClosePrice:     float64(closePrice) / divisor,
		TotalBuyQty:    totalBuyQty,
		TotalSellQty:   totalSellQty,
		AvgTradedPrice: float64(avgTradedPrice) / divisor,
	}, nil
}

func parseSnapQuotePacket(data []byte) (*MarketTick, error) {
	// Contains everything from Quote, plus more
	tick, err := parseQuotePacket(data[0:123])
	if err != nil {
		return nil, err
	}

	// Extra SnapQuote fields start at 147 (after best 5 data which is 200 bytes)
	// Best 5 Data: 147 to 347 (200 bytes)
	// Upper Circuit: 347
	upperCircuit := int64(binary.LittleEndian.Uint64(data[347:355]))
	lowerCircuit := int64(binary.LittleEndian.Uint64(data[355:363]))
	high52 := int64(binary.LittleEndian.Uint64(data[363:371]))
	low52 := int64(binary.LittleEndian.Uint64(data[371:379]))

	// Assuming exchange type is same as initial byte
	divisor := 100.0
	if data[1] == 13 {
		divisor = 10000000.0
	}

	tick.UpperCircuit = float64(upperCircuit) / divisor
	tick.LowerCircuit = float64(lowerCircuit) / divisor
	tick.High52Week = float64(high52) / divisor
	tick.Low52Week = float64(low52) / divisor

	return tick, nil
}

func mathFloat64frombits(b uint64) float64 {
	return math.Float64frombits(b)
}
