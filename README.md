# malgova
*warning :* "**work-in-progress**"

Algo backtest go-module, to help with writing day trading strategies for NSE Level 1 / Level 2 datasets
this go-module uses the kstreamdb for tick-data, https://github.com/sivamgr/kstreamdb
For recording market-data using zerodha Kite API, refer to kbridge tool available at, https://github.com/sivamgr/kbridge



# test
```console
C:\source\repo>git clone https://github.com/sivamgr/malgova.git
C:\source\repo>cd malgova
C:\source\repo\kbridge>go get -u
C:\source\repo\kbridge>go test
```

# go get
```console
C:\source\repo\kbridge>go get github.com/sivamgr/malgova
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

// emaCrossStrategy AlgoStrategy
type emaCrossStrategy struct {
	symbol string
	cs1m   *malgova.CandlesData
}

// Setup method, should return list of symbols it need to subscribe for tickdata
func (a *emaCrossStrategy) Setup(symbol string, b *malgova.Book) []string {
	symbolsToSubscribe := make([]string, 0)
	a.symbol = symbol
	a.cs1m = malgova.NewCandlesData(60)
    
    // add one or more symbols as needed by strategy
	symbolsToSubscribe = append(symbolsToSubscribe, symbol)
	b.AllocateCash(10000)
	return symbolsToSubscribe
}

// OnTick Method
func (a *emaCrossStrategy) OnTick(t kstreamdb.TickData, b *malgova.Book) {
	if t.TradingSymbol == a.symbol {
		a.cs1m.Update(t)
	}
}

// OnPeriodic method
func (a *emaCrossStrategy) OnPeriodic(t time.Time, b *malgova.Book) {
	if a.cs1m.HasChanged(t) && len(a.cs1m.Close) > 15 {
		ltp := a.cs1m.LTP
		ma1 := talib.Sma(a.cs1m.High, 15)
		ma2 := talib.Ema(a.cs1m.Close, 15)
		ma3 := talib.Sma(a.cs1m.Low, 15)
		if b.IsBookClean() && talib.Crossover(ma2, ma1) {
			//fmt.Printf("[%v] Buy @ %.2f\n", t, ltp)
			b.Buy(int(b.Cash / ltp))
		}
		if b.InPosition() && talib.Crossunder(ma2, ma3) {
			//fmt.Printf("[%v] Sell @ %.2f\n", t, ltp)
			b.Sell(b.Position)
		}
	}
}

// OnClose method
func (a *emaCrossStrategy) OnClose(b *malgova.Book) {
	b.Exit()
}


func main() {
	db := kstreamdb.SetupDatabase("/home/pi/data-kbridge/data/")
	bt := malgova.BacktestEngine{}
	bt.RegisterAlgo(emaCrossStrategy{})
	bt.Run(&db, nil)
}

```
