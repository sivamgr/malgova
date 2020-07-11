package malgova

import (
	"reflect"
	"time"

	"github.com/sivamgr/kstreamdb"
)

type btAlgoRunner struct {
	algoName            string
	symbol              string
	ptr                 reflect.Value
	ainterface          interface{}
	strategy            AlgoStrategy
	book                Book
	watch               []string
	enable              bool
	lastTick            kstreamdb.TickData
	queueTick           []kstreamdb.TickData
	utcLastPeriodicCall int64
	orders              []orderEntry
}

func (a *btAlgoRunner) ID() string {
	return a.algoName + "::" + a.symbol
}

func (a *btAlgoRunner) queue(t kstreamdb.TickData) {
	if a.enable {
		a.queueTick = append(a.queueTick, t)
	}
}

func (a *btAlgoRunner) resetQueue() {
	a.queueTick = make([]kstreamdb.TickData, 0, len(a.watch)*24000)
}

func (a *btAlgoRunner) run() {
	if a.enable {
		a.strategy.OnDayStart(&a.book)
		for _, t := range a.queueTick {
			a.checkClock(t.Timestamp)
			a.handleTick(t)
		}
		a.strategy.OnDayEnd(&a.book)
		a.resetQueue()
		//fmt.Printf("P/L %9.2f | Trades %3d | %s\n", a.book.Cash-a.book.CashAllocated, a.book.OrderCount, a.ID())
	}
}

func (a *btAlgoRunner) exit() {
	if a.enable {
		a.strategy.OnClose(&a.book)
		a.handleBook()
	}
}

func (a *btAlgoRunner) checkClock(t time.Time) {
	utcNow := t.Unix()
	if a.utcLastPeriodicCall < utcNow {
		a.utcLastPeriodicCall = utcNow
		a.strategy.OnPeriodic(time.Unix(utcNow, 0), &a.book)
	}
}

func (a *btAlgoRunner) handleBook() {
	if a.book.IsOrderWaiting() {
		if a.book.IsMarketOrder {
			if a.book.PendingOrderQuantity > 0 {
				buyPrice := a.lastTick.Ask[0].Price
				if buyPrice <= 0 {
					buyPrice = a.lastTick.LastPrice
				}
				cost := buyPrice * float32(a.book.PendingOrderQuantity)
				a.book.Cash -= float64(cost)
				a.book.Position += a.book.PendingOrderQuantity
				// add trade trade ledger
				a.orders = append(a.orders, orderEntry{
					algoName: a.algoName,
					at:       a.lastTick.Timestamp,
					symbol:   a.symbol,
					qty:      a.book.PendingOrderQuantity,
					price:    float64(buyPrice),
				})

				a.book.PendingOrderQuantity = 0
				a.book.OrderCount++
			} else if a.book.PendingOrderQuantity < 0 {
				sellPrice := a.lastTick.Bid[0].Price
				if sellPrice <= 0 {
					sellPrice = a.lastTick.LastPrice
				}

				cost := sellPrice * float32(a.book.PendingOrderQuantity)
				a.book.Cash -= float64(cost)
				a.book.Position += a.book.PendingOrderQuantity
				// add trade trade ledger
				a.orders = append(a.orders, orderEntry{
					algoName: a.algoName,
					at:       a.lastTick.Timestamp,
					symbol:   a.symbol,
					qty:      a.book.PendingOrderQuantity,
					price:    float64(sellPrice),
				})

				a.book.PendingOrderQuantity = 0
				a.book.OrderCount++
			}
		} else {
			if a.book.PendingOrderQuantity > 0 {
				if a.lastTick.LastPrice <= float32(a.book.PendingOrderPrice) {
					cost := a.book.PendingOrderPrice * float64(a.book.PendingOrderQuantity)
					a.book.Cash -= float64(cost)
					a.book.Position += a.book.PendingOrderQuantity
					// add trade trade ledger
					a.orders = append(a.orders, orderEntry{
						algoName: a.algoName,
						at:       a.lastTick.Timestamp,
						symbol:   a.symbol,
						qty:      a.book.PendingOrderQuantity,
						price:    float64(a.book.PendingOrderPrice),
					})
					a.book.PendingOrderQuantity = 0
					a.book.OrderCount++
				}
			} else if a.book.PendingOrderQuantity < 0 {
				if a.lastTick.LastPrice >= float32(a.book.PendingOrderPrice) {
					cost := a.book.PendingOrderPrice * float64(a.book.PendingOrderQuantity)
					a.book.Cash -= float64(cost)
					a.book.Position += a.book.PendingOrderQuantity
					// add trade trade ledger
					a.orders = append(a.orders, orderEntry{
						algoName: a.algoName,
						at:       a.lastTick.Timestamp,
						symbol:   a.symbol,
						qty:      a.book.PendingOrderQuantity,
						price:    float64(a.book.PendingOrderPrice),
					})
					a.book.PendingOrderQuantity = 0
					a.book.OrderCount++
				}
			}
		}
	}
}

func (a *btAlgoRunner) handleTick(t kstreamdb.TickData) {
	if (a.symbol == t.TradingSymbol) && t.IsTradable {
		a.lastTick = t
		a.handleBook()
	}
	a.strategy.OnTick(t, &a.book)
}

func (a *btAlgoRunner) popOrders() []orderEntry {
	orders := a.orders
	a.orders = make([]orderEntry, 0)
	return orders
}

func newAlgoInstance(algoType reflect.Type, symbol string) *btAlgoRunner {
	a := new(btAlgoRunner)
	a.algoName = algoType.Name()
	a.symbol = symbol
	a.book = Book{}
	a.ptr = reflect.New(algoType)
	a.strategy = a.ptr.Interface().(AlgoStrategy)
	a.watch = a.strategy.Setup(symbol, &a.book)
	a.enable = len(a.watch) > 0
	a.utcLastPeriodicCall = 0

	if a.enable {
		// prealloc queue
		a.resetQueue()
	}
	a.orders = make([]orderEntry, 0)
	//fmt.Printf("%+v %+v %+v \n", a.ptr, reflect.TypeOf(a.ptr), a.ptr.Interface().(AlgoStrategy))
	return a
}
