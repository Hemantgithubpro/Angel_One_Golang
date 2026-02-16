package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	// "github.com/joho/godotenv"
)

type mode string

const (
	ltpMode  mode = "LTP"
	ohlcMode mode = "OHLC"
	fullMode mode = "FULL"
)

type MarketItem struct {
	Exchange      string  `json:"exchange"`
	TradingSymbol string  `json:"tradingSymbol"`
	SymbolToken   string  `json:"symbolToken"`
	Ltp           float64 `json:"ltp"`
}

type QuoteResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	Errorcode string `json:"errorcode"`
	Data      struct {
		Fetched   []MarketItem `json:"fetched"`
		// Unfetched []string     `json:"unfetched"`
	} `json:"data"`
}

func getMarketData(apikey string, jwtToken string, exchange ExchangeType, symboltoken string, currentmode mode) {

	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/market/v1/quote/"
	method := "POST"

	payload := strings.NewReader(`{
      "mode": "` + string(currentmode) + `",
	  "exchangeTokens": {
		"` + string(exchange) + `": ["` + symboltoken + `"]
		}
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
		fmt.Println(err)
		return
	}
	fmt.Println("Response Status:", res.Status)
	fmt.Println("Response Body:", string(body))

	// Saving the response to CSV (optional)
	// var response QuoteResponse
	// if err := json.Unmarshal(body, &response); err != nil {
	// 	fmt.Println("JSON Unmarshal error:", err)
	// 	return
	// }

	// if !response.Status {
	// 	fmt.Println("API returned error:", response.Message)
	// 	return
	// }

	// if err := saveToCSV("market_data.csv", response.Data.Fetched); err != nil {
	// 	fmt.Println("Error saving CSV:", err)
	// 	return
	// }

	// fmt.Println("Saved response to market_data.csv")
}

func saveToCSV(filename string, data []MarketItem) error {
	csvFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("file create error: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write Header
	header := []string{"Exchange", "TradingSymbol", "SymbolToken", "LTP"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("CSV Header write error: %v", err)
	}

	for _, item := range data {
		record := []string{
			item.Exchange,
			item.TradingSymbol,
			item.SymbolToken,
			strconv.FormatFloat(item.Ltp, 'f', 2, 64),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("CSV Row write error: %v", err)
		}
	}

	return nil
}

func getMarketDataofMore(apikey string, jwtToken string, exchangeTokens map[string][]string, currentmode mode) {

	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/market/v1/quote/"
	method := "POST"

	requestBody := map[string]interface{}{
		"mode":           currentmode,
		"exchangeTokens": exchangeTokens,
	}

	jsonPayload, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return
	}

	payload := strings.NewReader(string(jsonPayload))

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
		fmt.Println(err)
		return
	}
	fmt.Println("Response Status:", res.Status)
	fmt.Println("Response Body:", string(body))

	// Saving the response to CSV (optional)
	// var response QuoteResponse
	// if err := json.Unmarshal(body, &response); err != nil {
	// 	fmt.Println("JSON Unmarshal error:", err)
	// 	return
	// }

	// if !response.Status {
	// 	fmt.Println("API returned error:", response.Message)
	// 	return
	// }

	// if err := saveToCSV("market_data.csv", response.Data.Fetched); err != nil {
	// 	fmt.Println("Error saving CSV:", err)
	// 	return
	// }

	// fmt.Println("Saved response to market_data.csv")
}

func marketDatatoDB(apikey string, jwtToken string, exchange ExchangeType, symboltoken string, currentmode mode) {

	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/market/v1/quote/"
	method := "POST"

	payload := strings.NewReader(`{
      "mode": "` + string(currentmode) + `",
	  "exchangeTokens": {
		"` + string(exchange) + `": ["` + symboltoken + `"]
		}
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
		fmt.Println(err)
		return
	}
	// fmt.Println("Response Status:", res.Status)
	// fmt.Println("Response Body:", string(body))

	// Saving the response to CSV (optional)
	var response QuoteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return
	}

	if !response.Status {
		fmt.Println("API returned error:", response.Message)
		return
	}

	if err := saveToCSV("market_data.csv", response.Data.Fetched); err != nil {
		fmt.Println("Error saving CSV:", err)
		return
	}

	fmt.Println("Saved response to market_data.csv")
}
