package malgova

import (
	"fmt"
	"time"
)

type tradeEntry struct {
	algoName string
	symbol   string
	at       time.Time
	qty      int
	price    float64
}

func (t tradeEntry) String() string {
	return fmt.Sprintf("%12s | %15s | %s | %4d | %9.2f", t.algoName, t.symbol, t.at.Format("2006/01/02 15:04:05"), t.qty, t.price)
}
