package malgova

import "time"

type tradeEntry struct {
	algoName string
	symbol   string
	at       time.Time
	qty      int
	price    float64
}
