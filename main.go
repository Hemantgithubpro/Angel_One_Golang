package main

import (
	// "fmt"
	"context"
	"log"
	"time"
)

func main() {
	apikey, jwtToken, clientId, feedToken, err := getCredentials()
	if err != nil {
		log.Printf("Error getting credentials: %v", err)
	}

	if jwtToken == "" || apikey == "" || clientId == "" || feedToken == "" {
		log.Fatal("Missing required environment variables: jwt_token, API_KEY, CLIENT_ID, feed_token")
	}

	// // fmt.Println("apikey: ",apikey)
	// // fmt.Println("jwttoken: ",jwtToken)
	// exchange := NSE
	// symboltoken := "99926000"
	// // getCandleData(apikey, jwtToken, exchange, symboltoken, ThreeMin, "2025-01-01 00:00", "2026-02-09 00:00")

	// // getMarketData(apikey, jwtToken, MCX, "467013", ltpMode)

	// // for i := 0; ; i++ {
	// // 	fmt.Println(marketDatatoDB(apikey, jwtToken, exchange, symboltoken, ltpMode))
	// // 	// Simulate a delay between API calls (e.g., 300ms)
	// // 	time.Sleep(300 * time.Millisecond)
	// // }

	// ticker := time.NewTicker(300 * time.Millisecond)
	// defer ticker.Stop()

	// // inFlight := false

	// // for range ticker.C {
	// // 	if inFlight {
	// // 		continue
	// // 	}
	// // 	inFlight = true

	// // 	go func() {
	// // 		fmt.Println(marketDatatoDB(apikey, jwtToken, exchange, symboltoken, ltpMode))
	// // 		inFlight = false
	// // 	}()
	// // }
	// for range ticker.C {
	// 	go func() {
	// 		fmt.Println(marketDatatoDB(apikey, jwtToken, exchange, symboltoken, ltpMode))
	// 	}()
	// }

	// --- DB & Buffer Setup ---
	db, err := NewDatabase()
	if err != nil {
		log.Printf("Warning: Database connection failed (continuing without DB): %v", err)
	} else {
		defer db.Close()
		log.Println("Database connected.")
		if err := db.InitSchema(context.Background()); err != nil {
			log.Printf("Warning: Failed to init schema: %v", err)
		}
	}

	buffer := NewTickBuffer()

	// Flush buffer to DB every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ticks := buffer.Flush()
			if len(ticks) > 0 {
				log.Printf("Flushing %d ticks to DB...", len(ticks))
				if db != nil {
					if err := db.BulkInsert(context.Background(), ticks); err != nil {
						log.Printf("Error inserting ticks: %v", err)
					}
				}
			}
		}
	}()

	// Start WebSocket Connection
	tokens := []TokenInfo{
		// {ExchangeType: 1, Tokens: []string{"99926000","2885"}},
		// {ExchangeType: 3, Tokens: []string{"99919000"}},
		// {ExchangeType: 2, Tokens: []string{"64862"}},
		{ExchangeType: 1, Tokens: []string{"2885"}},
	}
	websocketConnectiontoDB(jwtToken, apikey, clientId, feedToken, 2, tokens, buffer)

	// trydb()

	// websocketprint(jwtToken, apikey, clientId, feedToken, 2, tokens)
}
