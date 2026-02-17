package main

import "time"

// MarketTick represents a unified structure for all feed modes (LTP, Quote, SnapQuote)
type MarketTick struct {
	Token          string
	ExchangeType   int
	Timestamp      time.Time
	LTP            float64
	Volume         int64
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	ClosePrice     float64
	TotalBuyQty    float64
	TotalSellQty   float64
	AvgTradedPrice float64
	UpperCircuit   float64 // Only used in SnapQuote
	LowerCircuit   float64 // Only used in SnapQuote
	High52Week     float64 // Only used in SnapQuote
	Low52Week      float64 // Only used in SnapQuote
}
