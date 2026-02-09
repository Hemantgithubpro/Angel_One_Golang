//go:build websocket_example

// Simple Websocket connection and subscription example. make http request first to get tokens

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Errorf("error loading .env file: %v", err)
	}
	// Retrieve credentials from environment variables
	jwt_token := os.Getenv("jwt_token")
	api_key := os.Getenv("API_KEY")
	client_id := os.Getenv("CLIENT_ID")
	feed_token := os.Getenv("feed_token")

	if jwt_token == "" || api_key == "" || client_id == "" || feed_token == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_ID, FEED_TOKEN")
	}

	websocketConnection1(jwt_token, api_key, client_id, feed_token, 1, TokenInfo{
		ExchangeType: 3,
		Tokens:       []string{"99919000"},
	})
}

func websocketConnection1(jwt_token string, api_key string, client_id string, feed_token string, mode int, token TokenInfo) {
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
			Mode: mode, // LTP Mode
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

			if messageType == websocket.TextMessage {
				log.Printf("Received Text: %s", string(message))
			} else if messageType == websocket.BinaryMessage {
				parseBinaryResponse(message)
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

func parseBinaryResponse(data []byte) {
	if len(data) == 0 {
		return
	}

	mode := data[0]
	// exchangeType := data[1]
	// To distinguish currency for divisor, we might need exchangeType.
	// 1=nse_cm, 2=nse_fo, 3=bse_cm, 4=bse_fo, 5=mcx_fo, 7=ncx_fo, 13=cde_fo
	// For simplicity using 100.0 divisor. Real implementation should check ExchangeType 13.

	switch mode {
	case 1: // LTP Mode
		if len(data) != 51 {
			log.Printf("Invalid LTP packet size: %d", len(data))
			return
		}
		parseLTPPacket(data)
	case 2: // Quote Mode
		if len(data) != 123 {
			log.Printf("Invalid Quote packet size: %d", len(data))
			return
		}
		parseQuotePacket(data)
	case 3: // Snap Quote Mode
		if len(data) != 379 {
			log.Printf("Invalid SnapQuote packet size: %d", len(data))
			return
		}
		parseSnapQuotePacket(data)
	default:
		log.Printf("Unknown Subscription Mode: %d", mode)
	}
}

func parseLTPPacket(data []byte) {
	exchangeType := data[1]
	token := string(bytes.Trim(data[2:27], "\x00"))
	seqNum := int64(binary.LittleEndian.Uint64(data[27:35]))
	exchangeTime := int64(binary.LittleEndian.Uint64(data[35:43]))
	ltp := int64(binary.LittleEndian.Uint64(data[43:51]))

	divisor := 100.0
	if exchangeType == 13 {
		divisor = 10000000.0
	}
	realLTP := float64(ltp) / divisor
	tm := time.UnixMilli(exchangeTime)

	fmt.Printf("[LTP] Token: %s | Exch: %d | Time: %s | Price: %.2f | Seq: %d\n",
		token, exchangeType, tm.Format("15:04:05.000"), realLTP, seqNum)
}

func parseQuotePacket(data []byte) {
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

	fmt.Printf("[QUOTE] Token: %s | Time: %s | LTP: %.2f | Open: %.2f | High: %.2f | Low: %.2f | Close: %.2f | Vol: %d | BuyQ: %.0f | SellQ: %.0f | ATP: %.2f\n",
		token, tm.Format("15:04:05"),
		float64(ltp)/divisor,
		float64(openPrice)/divisor,
		float64(highPrice)/divisor,
		float64(lowPrice)/divisor,
		float64(closePrice)/divisor,
		volTraded, totalBuyQty, totalSellQty, float64(avgTradedPrice)/divisor)
}

func parseSnapQuotePacket(data []byte) {
	// Contains everything from Quote, plus more
	// Just reusing Quote parsing for the first part could work, but let's just parse the extra fields for now
	// or treat it similarly.
	// For brevity, just calling parseQuotePacket which will print the first part.
	// Note: In real app, we'd extract the common parsing logic.

	parseQuotePacket(data[0:123]) // Print the Quote part

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

	fmt.Printf("\t[SNAP] Upper: %.2f | Lower: %.2f | 52High: %.2f | 52Low: %.2f\n",
		float64(upperCircuit)/divisor, float64(lowerCircuit)/divisor,
		float64(high52)/divisor, float64(low52)/divisor)
}

func mathFloat64frombits(b uint64) float64 {
	return math.Float64frombits(b)
}
