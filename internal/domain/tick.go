package domain

import "time"

type Source string

const (
	SourceWebsocket Source = "websocket"
	SourcePoller    Source = "poller"
)

type Tick struct {
	Source         Source
	Token          string
	Exchange       string
	ExchangeType   int
	EventTime      time.Time
	ReceivedAt     time.Time
	LTP            float64
	Volume         int64
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	ClosePrice     float64
	TotalBuyQty    float64
	TotalSellQty   float64
	AvgTradedPrice float64
	UpperCircuit   float64
	LowerCircuit   float64
	High52Week     float64
	Low52Week      float64
	TradingSymbol  string
	LastTradedQty  int64
}
