//go:build websocket_example

// Simple Websocket connection and subscription example. make http request first to get tokens

package main

import (
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"fmt"
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
	client_code := os.Getenv("CLIENT_ID")
	feed_token := os.Getenv("feed_token")

	if jwt_token == "" || api_key == "" || client_code == "" || feed_token == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_CODE, FEED_TOKEN")
	}

	g(jwt_token, api_key, client_code, feed_token)
}

func g(jwt_token string, api_key string, client_code string, feed_token string) {
	// --- STEP 1: Websocket Connection ---
	log.Println("Step 1: Connecting to WebSocket...")

	// Set up the interrupt channel to handle Ctrl+C gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create request headers containing the token
	headers := http.Header{}
	headers.Add("Authorization", jwt_token)
	headers.Add("x-api-key", api_key)
	headers.Add("x-client-code", client_code)
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
			Mode: 1, // LTP Mode
			TokenList: []TokenInfo{
				{
					ExchangeType: 1, // NSE
					Tokens:       []string{"99926000"}, // Nifty 50
				},
				
			},
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
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			// In a real app, you would parse the binary message here
			
			log.Printf("Received message of %d bytes", len(message))
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
