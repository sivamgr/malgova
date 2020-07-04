package malgova

import (
	"log"
	"reflect"

	"github.com/sivamgr/kstreamdb"
)

// BacktestEngine struct
type BacktestEngine struct {
	algos  []reflect.Type
	orders []orderEntry
	scores []AlgoScore
}

// RegisterAlgo BacktestEngine
func (bt *BacktestEngine) RegisterAlgo(a interface{}) {
	if bt.algos == nil {
		bt.algos = make([]reflect.Type, 0)
	}
	bt.algos = append(bt.algos, reflect.TypeOf(a))
}

// Run BacktestEngine
func (bt *BacktestEngine) Run(feed *kstreamdb.DB, oms OrderManager) {
	// Load All Data into memory
	dates, _ := feed.GetDates()
	bt.orders = make([]orderEntry, 0)
	for _, dt := range dates {
		log.Printf("%s - Loading data into memory\n", dt.Format("20060102"))
		data, _ := feed.LoadDataForDate(dt)
		log.Printf("%s - %d ticks loaded\n", dt.Format("20060102"), len(data))
		dayRunner := btDayRunner{}
		dayRunner.run(bt.algos, data)
		// merge the trade ledger
		if len(dayRunner.orders) > 0 {
			bt.orders = append(bt.orders, dayRunner.orders...)
		}
		log.Printf("%s - completed run\n", dt.Format("20060102"))
	}

	// generate scores for algo runs
	bt.scores = calculateAlgoScores(bt.orders)
}

// Scores returns the scores calculated
func (bt *BacktestEngine) Scores() []AlgoScore {
	return bt.scores
}
