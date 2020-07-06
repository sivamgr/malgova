package malgova

import (
	"reflect"
	"time"

	"github.com/cskr/pubsub"
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
	utcLastPeriodicCall int64
	orders              []orderEntry
}

func (a *btAlgoRunner) ID() string {
	return a.algoName + "::" + a.symbol
}

func (a *btAlgoRunner) processTick(pubsub *pubsub.PubSub) {
	ch := pubsub.Sub(a.watch...)
	for msg := range ch {
		t := msg.(kstreamdb.TickData)
		a.checkClock(t.Timestamp)
		a.handleTick(t)
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

func newAlgoInstance(algoType reflect.Type, symbol string, pubsub *pubsub.PubSub) *btAlgoRunner {
	a := new(btAlgoRunner)
	a.algoName = algoType.Name()
	a.symbol = symbol
	a.book = Book{}
	a.ptr = reflect.New(algoType)
	a.strategy = a.ptr.Interface().(AlgoStrategy)
	a.watch = a.strategy.Setup(symbol, &a.book)
	a.enable = len(a.watch) > 0
	a.utcLastPeriodicCall = 0

	a.orders = make([]orderEntry, 0)
	if a.enable {
		// process ticks
		go a.processTick(pubsub)
	}
	//fmt.Printf("%+v %+v %+v \n", a.ptr, reflect.TypeOf(a.ptr), a.ptr.Interface().(AlgoStrategy))
	return a
}
