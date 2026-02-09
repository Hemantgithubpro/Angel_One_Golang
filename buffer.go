
package main

import (
	"sync"
)

type TickBuffer struct {
	mu     sync.Mutex
	buffer []MarketTick
}

func NewTickBuffer() *TickBuffer {
	return &TickBuffer{
		buffer: make([]MarketTick, 0, 1000),
	}
}

func (tb *TickBuffer) Add(tick MarketTick) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.buffer = append(tb.buffer, tick)
}

func (tb *TickBuffer) Flush() []MarketTick {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if len(tb.buffer) == 0 {
		return nil
	}

	data := tb.buffer
	tb.buffer = make([]MarketTick, 0, cap(data))
	return data
}
