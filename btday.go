package malgova

import (
	"reflect"
	"sync"

	"github.com/sivamgr/kstreamdb"
)

// btDayRunner struct
type btDayRunner struct {
	algos       []reflect.Type
	tickManager map[string]*btTickManager
	algoRunner  map[string]*btAlgoRunner
	trades      []tradeEntry
}

func (bt *btDayRunner) instantiateAllAlgosForSymbol(symbol string) {
	//spawn algos for symbol

	for _, a := range bt.algos {
		pAlgo := newAlgoInstance(a, symbol)
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

// worker for concurrent algo execution
func algoRunWorker(wg *sync.WaitGroup, algo *btAlgoRunner, bt *btDayRunner) {
	defer wg.Done()
	algo.run()
	// merge the trade ledger
	if len(algo.trades) > 0 {
		bt.trades = append(bt.trades, algo.trades...)
	}
}

//run day data against algos
func (bt *btDayRunner) run(algos []reflect.Type, ticks []kstreamdb.TickData) {
	// Feed Data to algos
	bt.algos = algos
	bt.tickManager = make(map[string]*btTickManager)
	bt.algoRunner = make(map[string]*btAlgoRunner)
	flagSymbolAlgoSetup := make(map[string]bool)
	bt.trades = make([]tradeEntry, 0)
	for _, t := range ticks {
		// instantiate algo runners if not instantiated already
		if t.IsTradable {
			if _, ok := flagSymbolAlgoSetup[t.TradingSymbol]; !ok {
				flagSymbolAlgoSetup[t.TradingSymbol] = true
				bt.instantiateAllAlgosForSymbol(t.TradingSymbol)
			}
		}

		// pass data to algos subscribed to this symbol
		if tickMgr, ok := bt.tickManager[t.TradingSymbol]; ok {
			for _, algoid := range tickMgr.observerAlgoIDs {
				pAlgo := bt.algoRunner[algoid]
				// queue tick for handling
				pAlgo.queue(t)
			}
		}
	}

	var wg sync.WaitGroup
	// run the runners
	for _, algo := range bt.algoRunner {
		wg.Add(1)
		go algoRunWorker(&wg, algo, bt)
	}

	wg.Wait()
}
