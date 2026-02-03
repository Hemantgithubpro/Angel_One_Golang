//go:build websocket_example

// Simple Websocket connection and subscription example. make http request first to get tokens

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Config represents target URLs
const (
	AuthURL      = "https://api.example.com/login"
	WebsocketURL = "wss://smartapisocket.angelone.in/smart-stream"
)

// 1. Structs for JSON parsing
type AuthRequest struct {
	Authorization string `json:"jwt_token"`
	Password      string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"access_token"`
}

type SubscriptionMessage struct {
	Action   string   `json:"action"`
	Channels []string `json:"channels"`
}

func g(jwt_token string, api_key string, client_code string, feed_token string) {
	// --- STEP 1: HTTP Authentication ---
	// log.Println("Step 1: Authenticating via HTTP...")
	// token, err := getAuthToken("myUser", "myPass123")
	// if err != nil {
	// 	log.Fatalf("Authentication failed: %v", err)
	// }
	// log.Printf("Got Token: %s...", token[:10]) // Log partial token for safety

	// --- STEP 2: WebSocket Connection ---
	log.Println("Step 2: Connecting to WebSocket...")

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
	conn, _, err := websocket.DefaultDialer.Dial(WebsocketURL, headers)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// --- STEP 3: Subscribe ---
	// Send a subscription message immediately after connecting
	subMsg := SubscriptionMessage{
		Action:   "subscribe",
		Channels: []string{"ticker-btc-usd", "trades"},
	}

	newjson := 
		{
     "correlationID": "abcde12345",
     "action": 1,
     "params": {
          "mode": 1,
          "tokenList": [
               {
                    "exchangeType": 1,
                    "tokens": [
                         "10626",
                         "5290"
                    ]
               },
               {
                    "exchangeType": 5,
                    "tokens": [
                         "234230",
                         "234235",
                         "234219"
                    ]
               }
          ]
     }
}
	

	err = conn.WriteJSON(newjson)
	if err != nil {
		log.Printf("Subscription failed: %v", err)
		return
	}
	log.Println("Subscribed to channels.")

	// --- STEP 4: Read Loop (Background) ---
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			log.Printf("Received: %s", message)
		}
	}()

	// --- STEP 5: Main Loop (Keep Alive / Shutdown) ---
	for {
		select {
		case <-done:
			return
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

// Helper function to perform the HTTP POST
func getAuthToken(user, pass string) (string, error) {
	// Prepare JSON payload
	payload := AuthRequest{Username: user, Password: pass}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Make HTTP POST request
	resp, err := http.Post(AuthURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}

	return authResp.Token, nil
}
