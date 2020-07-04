package malgova

import (
	"fmt"
	"testing"
	"time"

	"github.com/markcheno/go-talib"
	"github.com/sivamgr/kstreamdb"
)

// Momento AlgoStrategy
type Momento struct {
	symbol string
	cs1m   *CandlesData
}

// OnTick Method
func (a *Momento) OnTick(t kstreamdb.TickData, b *Book) {
	if t.TradingSymbol == a.symbol {
		a.cs1m.Update(t)
	}
}

// OnClose method
func (a *Momento) OnClose(b *Book) {
	b.Exit()
}

// OnPeriodic method
func (a *Momento) OnPeriodic(t time.Time, b *Book) {
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

// Setup method, should return list of symbols it need to subscribe for tickdata
func (a *Momento) Setup(symbol string, b *Book) []string {
	symbolsToSubscribe := make([]string, 0)
	a.symbol = symbol
	a.cs1m = NewCandlesData(60)
	symbolsToSubscribe = append(symbolsToSubscribe, symbol)
	b.AllocateCash(10000)
	return symbolsToSubscribe
}

func TestAStrategy(t *testing.T) {
	db := kstreamdb.SetupDatabase("/home/pi/data-kbridge/data/")
	bt := BacktestEngine{}
	bt.RegisterAlgo(Momento{})
	bt.Run(&db, nil)
	for _, trade := range bt.trades {
		fmt.Printf("%s\n", trade)
	}
}
