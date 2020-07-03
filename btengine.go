package malgova

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/sivamgr/kstreamdb"
)

// BacktestEngine struct
type BacktestEngine struct {
	algos               []reflect.Type
	tickManager         map[string]*tickDataManager
	algoInstances       map[string]*algoInstance
	flagSymbolAlgoSetup map[string]bool
	utcLastPeriodicCall int64
}

func getAlgoInstanceID(algoName string, symbol string) string {
	return algoName + "::" + symbol
}

type algoInstance struct {
	algoName   string
	symbol     string
	ptr        reflect.Value
	ainterface interface{}
	strategy   AlgoStrategy
	book       Book
	watch      []string
	enable     bool
	lastTick   kstreamdb.TickData
}

func (a *algoInstance) ID() string {
	return a.algoName + "::" + a.symbol
}

func newAlgoInstance(algoType reflect.Type, symbol string) *algoInstance {
	a := new(algoInstance)
	a.algoName = algoType.Name()
	a.symbol = symbol
	a.book = Book{}
	a.ptr = reflect.New(algoType)
	a.strategy = a.ptr.Interface().(AlgoStrategy)
	a.watch = a.strategy.Setup(symbol, &a.book)
	a.enable = len(a.watch) > 0
	//fmt.Printf("%+v %+v %+v \n", a.ptr, reflect.TypeOf(a.ptr), a.ptr.Interface().(AlgoStrategy))
	return a
}

type tickDataManager struct {
	observerAlgoIDs []string
}

func (s *tickDataManager) addObserver(algoID string) {
	if s.observerAlgoIDs == nil {
		s.observerAlgoIDs = make([]string, 0)
	}
	s.observerAlgoIDs = append(s.observerAlgoIDs, algoID)
}

// RegisterAlgo BacktestEngine
func (bt *BacktestEngine) RegisterAlgo(a interface{}) {
	if bt.algos == nil {
		bt.algos = make([]reflect.Type, 0)
	}
	bt.algos = append(bt.algos, reflect.TypeOf(a))
}

func (bt *BacktestEngine) instantiateAllAlgosForSymbol(symbol string) {
	//spawn algos for symbol

	for _, a := range bt.algos {
		pAlgo := newAlgoInstance(a, symbol)
		algoID := pAlgo.ID()
		bt.algoInstances[algoID] = pAlgo
		for _, w := range pAlgo.watch {
			if _, ok := bt.tickManager[w]; !ok {
				bt.tickManager[symbol] = new(tickDataManager)
			}
			bt.tickManager[symbol].addObserver(algoID)
		}
	}
}

// Run BacktestEngine
func (bt *BacktestEngine) Run(feed *kstreamdb.DB, oms OrderManager) {
	// Load All Data into memory
	dates, _ := feed.GetDates()
	for _, dt := range dates {
		log.Printf("Loading data for date %s into memory\n", dt.Format("20060102"))
		data, _ := feed.LoadDataForDate(dt)
		log.Printf("Date : loaded %d ticks\n", len(data))
		bt.testDayData(data)
		//break
	}

}

func (bt *BacktestEngine) checkClock(t time.Time) {
	utcNow := t.Unix()
	if bt.utcLastPeriodicCall < utcNow {
		bt.utcLastPeriodicCall = utcNow
		for _, algo := range bt.algoInstances {
			if algo.enable {
				algo.strategy.OnPeriodic(time.Unix(utcNow, 0), &algo.book)
			}
		}
	}
}

func (a *algoInstance) handleBook() {
	if a.book.IsOrderWaiting() {
		if a.book.IsMarketOrder {
			if a.book.PendingOrderQuantity > 0 {
				buyPrice := a.lastTick.Ask[0].Price
				cost := buyPrice * float32(a.book.PendingOrderQuantity)
				a.book.Cash -= float64(cost)
				a.book.Position += a.book.PendingOrderQuantity
				a.book.PendingOrderQuantity = 0
				a.book.OrderCount++
			} else if a.book.PendingOrderQuantity < 0 {
				sellPrice := a.lastTick.Bid[0].Price
				cost := sellPrice * float32(a.book.PendingOrderQuantity)
				a.book.Cash -= float64(cost)
				a.book.Position += a.book.PendingOrderQuantity
				a.book.PendingOrderQuantity = 0
				a.book.OrderCount++
			}
		} else {
			if a.book.PendingOrderQuantity > 0 {
				if a.lastTick.LastPrice <= float32(a.book.PendingOrderPrice) {
					cost := a.book.PendingOrderPrice * float64(a.book.PendingOrderQuantity)
					a.book.Cash -= float64(cost)
					a.book.Position += a.book.PendingOrderQuantity
					a.book.PendingOrderQuantity = 0
					a.book.OrderCount++
				}
			} else if a.book.PendingOrderQuantity < 0 {
				if a.lastTick.LastPrice >= float32(a.book.PendingOrderPrice) {
					cost := a.book.PendingOrderPrice * float64(a.book.PendingOrderQuantity)
					a.book.Cash -= float64(cost)
					a.book.Position += a.book.PendingOrderQuantity
					a.book.PendingOrderQuantity = 0
					a.book.OrderCount++
				}
			}
		}
	}
}

func (a *algoInstance) handleTick(t kstreamdb.TickData) {
	if (a.symbol == t.TradingSymbol) && t.IsTradable {
		a.lastTick = t
		a.handleBook()
	}
	a.strategy.OnTick(t, &a.book)
}

func (bt *BacktestEngine) testDayData(ticks []kstreamdb.TickData) {
	// Feed Data to algos
	bt.utcLastPeriodicCall = 0
	bt.tickManager = make(map[string]*tickDataManager)
	bt.algoInstances = make(map[string]*algoInstance)
	bt.flagSymbolAlgoSetup = make(map[string]bool)

	for _, t := range ticks {
		bt.checkClock(t.Timestamp)
		// instantiate algo if not instantiated already
		if t.IsTradable {
			if _, ok := bt.flagSymbolAlgoSetup[t.TradingSymbol]; !ok {
				bt.flagSymbolAlgoSetup[t.TradingSymbol] = true
				bt.instantiateAllAlgosForSymbol(t.TradingSymbol)
			}
		}

		// pass data to algos
		if _, ok := bt.tickManager[t.TradingSymbol]; ok {
			for _, algoid := range bt.tickManager[t.TradingSymbol].observerAlgoIDs {
				pAlgo := bt.algoInstances[algoid]
				// handle tick
				if pAlgo.enable {
					pAlgo.handleTick(t)
				}
			}
		}
	}
	for _, algo := range bt.algoInstances {
		if algo.enable {
			algo.strategy.OnClose(&algo.book)
			algo.handleBook()
			fmt.Printf("P/L %9.2f | Trades %3d | %s\n", algo.book.Cash-algo.book.CashAllocated, algo.book.OrderCount, algo.ID())
		}
	}
}
