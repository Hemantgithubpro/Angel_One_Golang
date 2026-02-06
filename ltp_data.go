package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func getLtpData(apikey string, jwtToken string, exchange string, tradingsymbol string, symboltoken string) {
	

	url := "https://apiconnect.angelone.in/rest/secure/angelbroking/order/v1/getLtpData"
	method := "POST"

	payload := strings.NewReader(`{
        "exchange": "` + exchange + `",
        "tradingsymbol": "` + tradingsymbol + `",
        "symboltoken": "` + symboltoken + `"
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
	fmt.Println("Response:", string(body))

}
