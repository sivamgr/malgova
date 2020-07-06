package malgova

import (
	"reflect"
	"time"

	"github.com/cskr/pubsub"
	"github.com/sivamgr/kstreamdb"
)

// btDayRunner struct
type btDayRunner struct {
	algos               []reflect.Type
	tickManager         map[string]*btTickManager
	algoRunner          map[string]*btAlgoRunner
	flagSymbolAlgoSetup map[string]bool
	orders              []orderEntry
	inProcPubSub        *pubsub.PubSub
}

func (bt *btDayRunner) instantiateAllAlgosForSymbol(symbol string) {
	//spawn algos for symbol

	for _, a := range bt.algos {
		pAlgo := newAlgoInstance(a, symbol, bt.inProcPubSub)
		algoID := pAlgo.ID()
		bt.algoRunner[algoID] = pAlgo
		for _, w := range pAlgo.watch {
			if _, ok := bt.tickManager[w]; !ok {
				bt.tickManager[symbol] = new(btTickManager)
			}
			bt.tickManager[symbol].addObserver(algoID)
		}
	}
}

func (bt *btDayRunner) setup(algos []reflect.Type) {
	bt.algos = algos
	bt.tickManager = make(map[string]*btTickManager)
	bt.algoRunner = make(map[string]*btAlgoRunner)
	bt.flagSymbolAlgoSetup = make(map[string]bool)
	// reset orders
	bt.orders = make([]orderEntry, 0)
	bt.inProcPubSub = pubsub.New(1)
}

func (bt *btDayRunner) exit() {
	for _, algo := range bt.algoRunner {
		algo.exit()
		// merge the trade ledger
		bt.orders = append(bt.orders, algo.popOrders()...)
	}
}

func (bt *btDayRunner) popOrders() []orderEntry {
	orders := bt.orders
	bt.orders = make([]orderEntry, 0)
	return orders
}

//run day data against algos
func (bt *btDayRunner) run(dt time.Time, ticks []kstreamdb.TickData) {

	for _, t := range ticks {
		// instantiate algo runners if not instantiated already
		if t.IsTradable {
			if _, ok := bt.flagSymbolAlgoSetup[t.TradingSymbol]; !ok {
				bt.flagSymbolAlgoSetup[t.TradingSymbol] = true
				bt.instantiateAllAlgosForSymbol(t.TradingSymbol)
			}
		}

		bt.inProcPubSub.Pub(t, t.TradingSymbol)
	}
}
