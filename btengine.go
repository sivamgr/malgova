package malgova

import (
	"log"
	"reflect"
	"sync"
	"time"

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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	var wg sync.WaitGroup

	for _, dt := range dates {
		log.Printf("[%s] Loading data", dt.Format("2006/01/02"))
		data, _ := feed.LoadDataForDate(dt)
		log.Printf("[%s] %d ticks loaded", dt.Format("2006/01/02"), len(data))
		wg.Wait()
		wg.Add(1)
		go func(d time.Time) {
			dayRunner.run(d, data)
			wg.Done()
			log.Printf("[%s] Completed", d.Format("2006/01/02"))
		}(dt)
	}
	wg.Wait()
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
