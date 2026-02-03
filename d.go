package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"strings"
)

func getCandleData(apikey string, jwtToken string, exchange string, symboltoken string, interval string, fromdate string, todate string) {
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
      "exchange": "` + exchange + `",
      "symboltoken": "` + symboltoken + `",
      "interval": "` + interval + `",
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

// func main() {
// 	getCandleData()
// }