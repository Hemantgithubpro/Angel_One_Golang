package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	OneMin     Interval = "ONE_MINUTE"
	ThreeMin   Interval = "THREE_MINUTE"
	FiveMin    Interval = "FIVE_MINUTE"
	TenMin     Interval = "TEN_MINUTE"
	FifteenMin Interval = "FIFTEEN_MINUTE"
	ThirtyMin  Interval = "THIRTY_MINUTE"
	OneHour    Interval = "ONE_HOUR"
	Oneday     Interval = "ONE_DAY"
)

type CandleResponse struct {
	Status    bool    `json:"status"`
	Message   string  `json:"message"`
	ErrorCode string  `json:"errorcode"`
	Data      [][]any `json:"data"`
}

type OIData struct {
	Time string  `json:"time"`
	Oi   float64 `json:"oi"`
}

type OIResponse struct {
	Status    bool     `json:"status"`
	Message   string   `json:"message"`
	ErrorCode string   `json:"errorcode"`
	Data      []OIData `json:"data"`
}

func getCandleData(apikey string, jwtToken string, exchange ExchangeType, symboltoken string, interval Interval, fromdate string, todate string) {

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

	var candleRes CandleResponse
	if err := json.NewDecoder(res.Body).Decode(&candleRes); err != nil {
		fmt.Println("JSON Decode error:", err)
		return
	}

	if !candleRes.Status {
		fmt.Println("API returned error:", candleRes.Message)
		return
	}

	csvFile, err := os.Create("data.csv")
	if err != nil {
		fmt.Println("File create error:", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write Header
	header := []string{"Timestamp", "Open", "High", "Low", "Close", "Volume"}
	if err := writer.Write(header); err != nil {
		fmt.Println("CSV Header write error:", err)
		return
	}

	for _, row := range candleRes.Data {
		record := make([]string, len(row))
		for i, col := range row {
			// Convert interface{} to string
			switch v := col.(type) {
			case float64:
				// Format float without scientific notation (-1 auto precision)
				record[i] = strconv.FormatFloat(v, 'f', -1, 64)
			default:
				record[i] = fmt.Sprintf("%v", col)
			}
		}
		if err := writer.Write(record); err != nil {
			fmt.Println("CSV Row write error:", err)
			return
		}
	}

	fmt.Println("Saved response to data.csv")

}

// Only for Derivatives (FNO) instruments
func getHistoricalOIData(apikey string, jwtToken string, exchange ExchangeType, symboltoken string, interval Interval, fromdate string, todate string) {
	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/historical/v1/getOIData"
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
	req.Header.Add("X-MACAddress", "00:00:00:00:00:00") // Replace with actual if needed
	req.Header.Add("X-PrivateKey", apikey)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer res.Body.Close()

	var oiRes OIResponse
	if err := json.NewDecoder(res.Body).Decode(&oiRes); err != nil {
		fmt.Println("JSON Decode error:", err)
		return
	}

	if !oiRes.Status {
		fmt.Println("API returned error:", oiRes.Message)
		return
	}

	csvFile, err := os.Create("oi_data.csv")
	if err != nil {
		fmt.Println("File create error:", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write Header
	header := []string{"Timestamp", "OpenInterest"}
	if err := writer.Write(header); err != nil {
		fmt.Println("CSV Header write error:", err)
		return
	}

	for _, row := range oiRes.Data {
		record := []string{
			row.Time,
			strconv.FormatFloat(row.Oi, 'f', -1, 64),
		}
		if err := writer.Write(record); err != nil {
			fmt.Println("CSV Row write error:", err)
			return
		}
	}

	fmt.Println("Saved OI response to oi_data.csv")
}
