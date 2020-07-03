package malgova

import (
	"time"

	"github.com/sivamgr/kstreamdb"
)

// Book struct
type Book struct {
	CashAllocated        float64
	Cash                 float64
	Position             int
	IsMarketOrder        bool
	PendingOrderQuantity int
	PendingOrderPrice    float64
	OrderCount           int
}

// OrderManager Interface
type OrderManager interface {
	PlaceLimitOrder(symbol string, qty int, price float64, a AlgoStrategy)
	PlaceMarketOrder(symbol string, qty int, price float64, a AlgoStrategy)
}

// Engine Interface
type Engine interface {
	RegisterAlgo(algo interface{})
	Run(feed *kstreamdb.DB, oms OrderManager)
	SubscribeChannel(Symbol string, a AlgoStrategy)
}

// AlgoStrategy Interface
type AlgoStrategy interface {
	Setup(symbol string, b *Book) []string
	OnTick(t kstreamdb.TickData, b *Book)
	OnPeriodic(t time.Time, b *Book) // Invokes every sec
	OnClose(b *Book)
}

/*
func getType(myvar interface{}) string {
	t := reflect.TypeOf(myvar)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}
	return t.Name()
}
*/
