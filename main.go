package main

import (
	// "context"
	// "fmt"
	"log"
	// "os"
	// "time"
	// "github.com/joho/godotenv"
)


func main() {
	apikey, jwtToken, clientId, feedToken, err := getCredentials()
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
	}

	if jwtToken == "" || apikey == "" || clientId == "" || feedToken == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_ID, feed_token")
	}

	exchange := NSE
	symboltoken := "99926000"
	getCandleData(apikey, jwtToken, exchange, symboltoken, ThreeMin, "2025-01-01 00:00", "2026-02-09 00:00")

	// --- DB & Buffer Setup ---
	// db, err := NewDatabase()
	// if err != nil {
	// 	log.Printf("Warning: Database connection failed (continuing without DB): %v", err)
	// } else {
	// 	defer db.Close()
	// 	log.Println("Database connected.")
	// 	if err := db.InitSchema(context.Background()); err != nil {
	// 		log.Printf("Warning: Failed to init schema: %v", err)
	// 	}
	// }

	// buffer := NewTickBuffer()

	// // Flush buffer to DB every 5 seconds
	// go func() {
	// 	ticker := time.NewTicker(5 * time.Second)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		ticks := buffer.Flush()
	// 		if len(ticks) > 0 {
	// 			log.Printf("Flushing %d ticks to DB...", len(ticks))
	// 			if db != nil {
	// 				if err := db.BulkInsert(context.Background(), ticks); err != nil {
	// 					log.Printf("Error inserting ticks: %v", err)
	// 				}
	// 			}
	// 		}
	// 	}
	// }()

	// // Start WebSocket Connection
	// websocketConnection1(jwtToken, apikey, clientId, feedToken, 1, TokenInfo{
	// 	ExchangeType: 3,
	// 	Tokens:       []string{"99919000"},
	// }, buffer)
}
