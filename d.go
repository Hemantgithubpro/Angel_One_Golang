package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"strings"
)

type ExchangeType string

const (
	NSE ExchangeType = "NSE" // National Stock Exchange
	NFO ExchangeType = "NFO" // NSE Futures & Options
	BSE ExchangeType = "BSE" // Bombay Stock Exchange
	BFO ExchangeType = "BFO" // BSE Futures & Options
	CDS ExchangeType = "CDS" // Currency Derivatives Segment
	MCX ExchangeType = "MCX" // Multi Commodity Exchange
)

type Interval string

const (
	OneMin   Interval = "ONE_MINUTE"
	ThreeMin Interval = "THREE_MINUTE"
	FiveMin  Interval = "FIVE_MINUTE"
	TenMin   Interval = "TEN_MINUTE"
	FifteenMin Interval = "FIFTEEN_MINUTE"
	ThirtyMin  Interval = "THIRTY_MINUTE"
	OneHour	Interval = "ONE_HOUR"
	Oneday	Interval = "ONE_DAY"
)



func getCandleData(apikey string, jwtToken string, exchange ExchangeType, symboltoken string, interval Interval, fromdate string, todate string) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// apikey := os.Getenv("API_KEY")
	// jwtToken := os.Getenv("jwt_token")

	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/historical/v1/getCandleData"
	method := "POST"

	payload := strings.NewReader(`{
      "exchange": "` + string(exchange) + `",
      "symboltoken": "` + symboltoken + `",
      "interval": "` + string(interval) + `",
      "fromdate": "` + fromdate + `",
      "todate": "` + todate + `"
 	}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println("Request creation error:", err)
		return
	}

	// Add headers
	req.Header.Add("Authorization", jwtToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-UserType", "USER")
	req.Header.Add("X-SourceID", "WEB")
	req.Header.Add("X-ClientLocalIP", "127.0.0.1")      // Replace with actual if needed
	req.Header.Add("X-ClientPublicIP", "198.168.0.1")   // Replace with actual if needed
	req.Header.Add("X-MACAddress", "00:0a:95:9d:68:16") // Replace with actual if needed
	req.Header.Add("X-PrivateKey", apikey)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	if err := os.WriteFile("response.json", body, 0644); err != nil {
		fmt.Println("File write error:", err)
		return
	}
	fmt.Println("Saved response to response.json")

}

// Only for Derivatives (FNO) instruments
// func getHistoricalOIData() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		fmt.Println("Error loading .env file")
// 		return
// 	}

// }

// func main() {
// 	getCandleData()
// }
