package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/pquerna/otp/totp"

	// "reflect"
	"strings"
	"time"
)

func getCredentials() (string, string, string, string, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %v", err)
	}
	// Read required environment variables
	mpin := os.Getenv("MPIN")
	clientId := os.Getenv("CLIENT_ID")
	totp_secret := os.Getenv("TOTP_SECRET")
	apikey := os.Getenv("API_KEY")

	// Validate env vars early
	if mpin == "" || clientId == "" || totp_secret == "" || apikey == "" {
		return "", "", "", "", fmt.Errorf("missing one or more required environment variables: MPIN, CLIENT_ID, TOTP_SECRET, API_KEY")
	}

	// Generate TOTP code (avoid shadowing the package name)
	totpCode, err := totp.GenerateCode(totp_secret, time.Now())
	if err != nil {
		panic(err)
	}
	url := "https://apiconnect.angelone.in/rest/auth/angelbroking/user/v1/loginByPassword"
	method := "POST"

	payload := strings.NewReader(`{
	"clientcode": "` + clientId + `",
	"password": "` + mpin + `",
	"totp":"` + totpCode + `",
	"state":"environment_variable"
	}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", "", "", "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-UserType", "USER")
	req.Header.Add("X-SourceID", "WEB")
	req.Header.Add("X-ClientLocalIP", "CLIENT_LOCAL_IP")
	req.Header.Add("X-ClientPublicIP", "CLIENT_PUBLIC_IP")
	req.Header.Add("X-MACAddress", "MAC_ADDRESS")
	// Use the real API key from environment
	req.Header.Add("X-PrivateKey", apikey)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", "", "", "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", "", "", err
	}

	var resp struct {
		Status    bool   `json:"status"`
		Message   string `json:"message"`
		Errorcode string `json:"errorcode"`
		Data      struct {
			JwtToken     string `json:"jwtToken"`
			RefreshToken string `json:"refreshToken"`
			FeedToken    string `json:"feedToken"`
			State        string `json:"state"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Println("json unmarshal error:", err)
		fmt.Println(string(body))
		return "", "", "", "", err
	}

	// fmt.Println(resp.Status)
	// fmt.Println(resp.Message)
	// fmt.Println(resp.Errorcode)
	// fmt.Println(resp.Data.JwtToken)
	// fmt.Println(resp.Data.RefreshToken)
	// fmt.Println(resp.Data.FeedToken)
	// fmt.Println(resp.Data.State)

	// Check API response for success
	if !resp.Status {
		return "", "", "", "", fmt.Errorf("login failed: %s", resp.Message)
	}
	jwtToken := string("Bearer " + resp.Data.JwtToken)
	feedToken := resp.Data.FeedToken
	if jwtToken == "" {
		return "", "", "", "", fmt.Errorf("login succeeded but jwtToken empty: %s", resp.Message)
	}

	return apikey, jwtToken, clientId, feedToken, nil

}
