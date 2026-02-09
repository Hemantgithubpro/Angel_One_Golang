package main

import (
	"fmt"
	// "net/http"
	"github.com/joho/godotenv"
	// "io"
	"os"
	// "strings"
	"time"
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
	// conn, err := wssConnect(jwtToken, apikey, clientCode, feedToken)
	// if err != nil {
	// 	fmt.Printf("Error connecting to WebSocket: %v\n", err)
	// 	return
	// }

	// Create streaming request
	// Using Nifty 50 token (99926000) on NSE (ExchangeType 1) as example
	// req := StreamRequest{
	// 	CorrelationID: "test_stream_123",
	// 	Action:        1, // Subscribe
	// 	Params: StreamParams{
	// 		Mode: 1, // LTP Mode
	// 		TokenList: []TokenInfo{
	// 			{
	// 				ExchangeType: 1,                    // NSE
	// 				Tokens:       []string{"99926000"}, // Nifty 50
	// 			},
	// 		},
	// 	},
	// }

	// fmt.Println("Starting stream...")

	// g(jwtToken, apikey, clientCode, feedToken)

	clientCode = clientCode + feedToken + apikey + jwtToken // just to avoid unused variable error

	//Historical data
	// exchange := NSE
	// symboltoken := "2885"
	// interval := OneHour
	// fromdate := "2026-02-01 00:00"
	// todate := "2026-02-02 23:59"
	// getCandleData(apikey,jwtToken,exchange,symboltoken,interval,fromdate,todate)

	// exchange := MCX
	// symboltoken := "467013"
	// interval := OneHour
	// fromdate := "2026-02-01 00:00"
	// todate := "2026-02-02 23:59"
	// getCandleData(apikey,jwtToken,exchange,symboltoken,interval,fromdate,todate)

	// exchange:= NFO
	// symboltoken := "48178"
	// interval := FifteenMin
	// fromdate := "2026-01-30 12:00"
	// todate := "2026-02-01 12:00"
	// getHistoricalOIData(apikey, jwtToken, exchange, symboltoken, interval, fromdate, todate)


	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	exchange := NFO
	symbolcode := "48236"
	for range ticker.C {
		getMarketData(apikey, jwtToken,exchange, symbolcode, ltpMode)
		// getMarketData(apikey, jwtToken,exchange, symbolcode, fullMode)
		// getMarketData(apikey, jwtToken,exchange, symbolcode, ohlcMode)
	}

	// exchangetokenmap := map[string][]string{
	// 	string(NSE): {"3045", "881"}, string(NFO): {"48236"},
	// }
	// getMarketDataofMore(apikey, jwtToken, exchangetokenmap, ltpMode)

}
