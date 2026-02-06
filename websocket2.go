// Web Socket Implementation in Go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func wssConnect(jwt_token string, api_key string, client_code string, feed_token string) (*websocket.Conn, error) {
	// wss://smartapisocket.angelone.in/smart-stream
	u := url.URL{Scheme: "wss", Host: "smartapisocket.angelone.in", Path: "/smart-stream"}
	header := make(map[string][]string)
	header["Authorization"] = []string{jwt_token}
	header["x-api-key"] = []string{api_key}
	header["x-client-code"] = []string{client_code}
	header["x-feed-token"] = []string{feed_token}

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Printf("WebSocket connection error: %v", err)
		if resp != nil {
			log.Printf("Handshake status: %d", resp.StatusCode)
			log.Printf("Response headers: %v", resp.Header)
		}
		return nil, err
	}
	log.Printf("WebSocket connected to %s", u.String())
	return c, nil
}

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

func subscribeAndStream(conn *websocket.Conn, req StreamRequest) error {
	defer conn.Close()

	// Send subscription request
	log.Printf("Sending subscription request...")
	if err := conn.WriteJSON(req); err != nil {
		log.Printf("Write error: %v", err)
		return err
	}

	// Start Heartbeat Ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Heartbeat Goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				// WriteMessage is not thread-safe by default in Gorilla websocket for concurrent writes,
				// but here main loop is reading and this routine is writing.
				// However, if we needed to write from multiple routines, we'd need a mutex.
				// Since we only write from here (and the initial subscription), it's mostly safe,
				// BUT strictly speaking concurrent WriteJSON/WriteMessage is not allowed.
				// For a simple script, this might be okay, or we should use a channel to coordinate writes.
				// Let's take the risk for a simple script or add a mutex if needed.
				// Actually, to be safe, let's just send. The initial write is before the loop starts.
				// So only this goroutine writes.
				err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					log.Printf("Error sending heartbeat: %v", err)
					return
				}
				log.Println("Sent Heartbeat: ping")
			}
		}
	}()

	// Stream Loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			return err
		}

		if messageType == websocket.TextMessage {
			log.Printf("Received Text Message: %s", string(message))
		} else if messageType == websocket.BinaryMessage {
			parseBinaryResponse(message)
		}
	}
}

func parseBinaryResponse(data []byte) {
	// Basic validation based on LTP packet size (51 bytes)
	// Byte 0: Subscription Mode
	if len(data) == 0 {
		return
	}

	subMode := data[0]

	switch subMode {
	case 1: // LTP Mode
		if len(data) != 51 {
			log.Printf("Invalid LTP packet size: %d", len(data))
			return
		}
		parseLTPPacket(data)
	case 2: // Quote Mode
		// Implement Quote parsing if needed
		log.Printf("Quote Mode packet received (size %d)", len(data))
	case 3: // Snap Quote Mode
		// Implement Snap Quote parsing if needed
		log.Printf("Snap Quote Mode packet received (size %d)", len(data))
	default:
		log.Printf("Unknown Subscription Mode: %d", subMode)
	}
}

func parseLTPPacket(data []byte) {
	// Field 1: Subscription Mode (byte) - already checked
	// subMode := data[0]

	// Field 2: Exchange Type (byte)
	exchangeType := data[1]

	// Field 3: Token (25 bytes)
	tokenBytes := data[2:27]
	// Trim null characters
	token := string(bytes.Trim(tokenBytes, "\x00"))

	// Field 4: Sequence Number (int64)
	seqNum := int64(binary.LittleEndian.Uint64(data[27:35]))

	// Field 5: Exchange Timestamp (int64)
	exchangeTime := int64(binary.LittleEndian.Uint64(data[35:43])) // Epoch milliseconds

	// Field 6: Last Traded Price (int64) - Size 8 bytes in spec
	ltpVal := int64(binary.LittleEndian.Uint64(data[43:51]))

	// Price conversion (divide by 100 for normal, 10000000.0 for currency)
	// Assuming non-currency for simplicity or check exchange type
	realLTP := float64(ltpVal) / 100.0

	// Convert timestamp to human readable
	tm := time.UnixMilli(exchangeTime)

	fmt.Printf("\n--- LTP Update ---\n")
	fmt.Printf("Token: %s (Exch: %d)\n", token, exchangeType)
	fmt.Printf("LTP: %.2f\n", realLTP)
	fmt.Printf("Time: %s\n", tm.Format("15:04:05.000"))
	fmt.Printf("Seq: %d\n", seqNum)
	fmt.Println("------------------")
}
