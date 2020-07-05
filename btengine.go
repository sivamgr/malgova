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
	dayRunner := btDayRunner{}
	dayRunner.setup(bt.algos)

	for _, dt := range dates {
		log.Printf("%s - Loading data into memory\n", dt.Format("20060102"))
		data, _ := feed.LoadDataForDate(dt)
		log.Printf("%s - %d ticks loaded\n", dt.Format("20060102"), len(data))
		dayRunner.run(data)
		// merge the trade ledger
		log.Printf("%s - completed run\n", dt.Format("20060102"))
	}
	dayRunner.exit()
	//pull the orders from the run
	bt.orders = dayRunner.popOrders()
	// analyze the orders and generate scores for algo
	bt.scores = calculateAlgoScores(bt.orders)
}

// Scores returns the scores calculated
func (bt *BacktestEngine) Scores() []AlgoScore {
	return bt.scores
}
