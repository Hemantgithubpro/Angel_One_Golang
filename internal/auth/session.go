package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
)

type Session struct {
	APIKey    string
	ClientID  string
	JWTToken  string
	FeedToken string
}

type Client struct {
	apiKey     string
	clientID   string
	mpin       string
	totpSecret string
	loginURL   string
	httpClient *http.Client
}

func NewClient(apiKey, clientID, mpin, totpSecret string) *Client {
	return &Client{
		apiKey:     apiKey,
		clientID:   clientID,
		mpin:       mpin,
		totpSecret: totpSecret,
		loginURL:   "https://apiconnect.angelone.in/rest/auth/angelbroking/user/v1/loginByPassword",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) SetLoginURL(url string) {
	c.loginURL = url
}

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) Login(ctx context.Context) (Session, error) {
	totpCode, err := totp.GenerateCode(c.totpSecret, time.Now())
	if err != nil {
		return Session{}, fmt.Errorf("generate totp: %w", err)
	}

	payload := map[string]string{
		"clientcode": c.clientID,
		"password":   c.mpin,
		"totp":       totpCode,
		"state":      "environment_variable",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Session{}, fmt.Errorf("marshal login payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.loginURL, bytes.NewReader(body))
	if err != nil {
		return Session{}, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-UserType", "USER")
	req.Header.Set("X-SourceID", "WEB")
	req.Header.Set("X-ClientLocalIP", "127.0.0.1")
	req.Header.Set("X-ClientPublicIP", "127.0.0.1")
	req.Header.Set("X-MACAddress", "00:00:00:00:00:00")
	req.Header.Set("X-PrivateKey", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Session{}, fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			JWTToken  string `json:"jwtToken"`
			FeedToken string `json:"feedToken"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Session{}, fmt.Errorf("decode login response: %w", err)
	}
	if !result.Status {
		return Session{}, fmt.Errorf("login failed: %s", result.Message)
	}

	return Session{
		APIKey:    c.apiKey,
		ClientID:  c.clientID,
		JWTToken:  "Bearer " + result.Data.JWTToken,
		FeedToken: result.Data.FeedToken,
	}, nil
}
