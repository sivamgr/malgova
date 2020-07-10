# malgova
*warning :* "**work-in-progress**"

Algo backtest go-module, to help with writing day trading strategies for NSE Level 1 / Level 2 datasets. This go-module uses the kstreamdb for tick-data, https://github.com/sivamgr/kstreamdb . For recording market-data using zerodha Kite API, refer to kbridge tool available at, https://github.com/sivamgr/kbridge


# go get
```console
go get github.com/sivamgr/malgova
```

# AlgoStrategy Interface

Algo strategies written in go should fully implement the malgova.AlgoStrategy interface as defined below

```go
// AlgoStrategy Interface
type AlgoStrategy interface {
	Setup(symbol string, b *Book) []string
	OnTick(t kstreamdb.TickData, b *Book)
	OnPeriodic(t time.Time, b *Book) // Invokes every sec
	OnClose(b *Book)
}
```
# Order Book

the Order-Book is passed to algo-strategies callback. Orders can be placed or position shall be exit through the methods exposed by the book


# Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/sivamgr/kstreamdb"
	"github.com/sivamgr/malgova"
)

// Momento AlgoStrategy
type Momento struct {
	symbol      string
	candles1min *malgova.CandlesData
}

// Setup method, should return list of symbols it need to subscribe for tickdata
func (a *Momento) Setup(symbol string, b *malgova.Book) []string {
	symbolsToSubscribe := make([]string, 0)
	a.symbol = symbol

	//set up data aggregation.
	a.candles1min = malgova.NewCandlesData(60)

	// add symbols needed to subscribe
	symbolsToSubscribe = append(symbolsToSubscribe, symbol)
	b.AllocateCash(10000)
	return symbolsToSubscribe
}
// OnDayStart method
func (a *Momento) OnDayStart(b *malgova.Book) {
}

// OnDayEnd method
func (a *Momento) OnDayEnd(b *malgova.Book) {
}


// OnTick Method
func (a *Momento) OnTick(t kstreamdb.TickData, b *malgova.Book) {
	if t.TradingSymbol == a.symbol {
		// update data aggregation on tick
		a.candles1min.Update(t)
	}
}

// OnPeriodic method
func (a *Momento) OnPeriodic(t time.Time, b *malgova.Book) {
	if a.cs1m.HasChanged(t) && len(a.cs1m.Close) > 15 {
		ltp := a.cs1m.LTP
		ma1 := talib.Sma(a.cs1m.High, 15)
		ma2 := talib.Ema(a.cs1m.Close, 15)
		ma3 := talib.Sma(a.cs1m.Low, 15)
		if b.IsBookClean() && talib.Crossover(ma2, ma1) {
			quantityToBuy := int(b.Cash / ltp)
			b.Buy(quantityToBuy)
		}
		if b.InPosition() && talib.Crossunder(ma2, ma3) {
			b.Exit()
		}
	}
}

// OnClose method
func (a *Momento) OnClose(b *malgova.Book) {
	b.Exit()
}

func main() {
	db := kstreamdb.SetupDatabase("/home/pi/test-data/")
	bt := malgova.BacktestEngine{}
	bt.RegisterAlgo(Momento{})
	bt.Run(&db, nil)
	for _, s := range bt.Scores() {
		fmt.Println(s)
	}
}

```
