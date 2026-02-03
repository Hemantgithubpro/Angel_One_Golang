package main

import (
	"fmt"
	// "net/http"
	"github.com/joho/godotenv"
	// "io"
	"os"
	// "strings"
)

func getCredentials() (string, string, string, string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", "", "", "", fmt.Errorf("error loading .env file: %v", err)
	}

	apikey := os.Getenv("API_KEY")
	jwtToken := os.Getenv("jwt_token")
	clientCode := os.Getenv("CLIENT_CODE")
	feedToken := os.Getenv("FEED_TOKEN")
	return apikey, jwtToken, clientCode, feedToken, nil
}

func main() {
	apikey, jwtToken, clientCode, feedToken, err := getCredentials()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Connect to WebSocket
	conn, err := wssConnect(jwtToken, apikey, clientCode, feedToken)
	if err != nil {
		fmt.Printf("Error connecting to WebSocket: %v\n", err)
		return
	}



	// Create streaming request
	// Using Nifty 50 token (99926000) on NSE (ExchangeType 1) as example
	req := StreamRequest{
		CorrelationID: "test_stream_123",
		Action:        1, // Subscribe
		Params: StreamParams{
			Mode: 1, // LTP Mode
			TokenList: []TokenInfo{
				{
					ExchangeType: 1,                    // NSE
					Tokens:       []string{"99926000"}, // Nifty 50
				},
			},
		},
	}

	fmt.Println("Starting stream...")
	err = subscribeAndStream(conn, req)
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
	}
}
