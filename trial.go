package main

import (
	"fmt"
	"net/http"
)

// Making a simple http request to demonstrate the use of net/http package
func get(url string) (*http.Response, error) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return resp, nil
}

// making http request with headers
func getWithHeaders(url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}


func trial(){
	url:= "http://example.com"
	// resp, err := get(url)
	// fmt.Println(resp, err)

	headers := map[string]string{
		"Content-Type":     "application/json",
		"X-ClientLocalIP":  "CLIENT_LOCAL_IP",
		"X-ClientPublicIP": "CLIENT_PUBLIC_IP",
		"X-MACAddress":     "MAC_ADDRESS",
		"Accept":           "application/json",
		"X-PrivateKey":     "api_key",
		"X-UserType":       "USER",
		"X-SourceID":       "WEB",
		"Authorization":    "Bearer <JWT_TOKEN>",
	}
	resp, err := getWithHeaders(url, headers)
	fmt.Println(resp, err)

}