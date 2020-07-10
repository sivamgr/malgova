package malgova

import (
	"fmt"

	"github.com/sivamgr/kstreamdb"
)

// PriceCell struct
type PriceCell struct {
	BidQuantityTaken uint32
	AskQuantityTaken uint32
	VolumeTraded     uint32
}

// OrderFlowMonitor struct
type OrderFlowMonitor struct {
	LastTick               kstreamdb.TickData
	TotalBidsQuantityTaken uint32
	TotalAsksQuantityTaken uint32
	TicksUpdated           uint32
	Prices                 map[uint32]*PriceCell
	Bids                   [5]kstreamdb.DepthItem
	Asks                   [5]kstreamdb.DepthItem
}

// NewOrderFlowMonitor creates a new orderflow monitor
func NewOrderFlowMonitor() *OrderFlowMonitor {
	pNew := new(OrderFlowMonitor)
	pNew.Prices = make(map[uint32]*PriceCell)
	return pNew
}

// GetPriceCell returns the price cell
func (r *OrderFlowMonitor) GetPriceCell(p float32) *PriceCell {
	k := uint32(p * 100)
	if _, ok := r.Prices[k]; !ok {
		r.Prices[k] = new(PriceCell)
	}
	return r.Prices[k]
}

// Update processes the tick
func (r *OrderFlowMonitor) Update(t kstreamdb.TickData) {
	if t.VolumeTraded > r.LastTick.VolumeTraded {
		r.TicksUpdated++
		ltq := t.LastTradedQuantity
		pCell := r.GetPriceCell(t.LastPrice)
		if (len(r.Bids) > 0) && (t.LastPrice <= r.Bids[0].Price) {
			pCell.BidQuantityTaken += ltq
			r.TotalBidsQuantityTaken += ltq
		} else if (len(r.Asks) > 0) && (t.LastPrice >= r.Bids[0].Price) {
			pCell.AskQuantityTaken += ltq
			r.TotalAsksQuantityTaken += ltq
		}

		pCell.VolumeTraded += ltq

		// Update with latest depth
		r.Bids = t.Bid
		r.Asks = t.Ask
	}

	r.LastTick = t
}

func (v *PriceCell) resetCounters() {
	v.AskQuantityTaken = 0
	v.BidQuantityTaken = 0
}

// Reset resets Bids/Asks
func (r *OrderFlowMonitor) Reset() {
	r.TotalAsksQuantityTaken = 0
	r.TotalBidsQuantityTaken = 0
	r.TicksUpdated = 0
	for _, v := range r.Prices {
		v.resetCounters()
	}
	r.LastTick = kstreamdb.TickData{}
}

// StringifyPriceCell Stringify the Price Bucket
func (r *OrderFlowMonitor) StringifyPriceCell(p float32) string {

	pCell := r.GetPriceCell(p)
	if pCell == nil {
		return ""
	}
	bidsQty := "        "
	bidsCount := "        "
	asksQty := "        "
	asksCount := "        "
	if (len(r.Bids) > 0) && (p <= r.Bids[0].Price) && (p >= r.Bids[len(r.Bids)-1].Price) {
		for _, bid := range r.Bids {
			if bid.Price == p {
				bidsCount = fmt.Sprintf("%8d", bid.Orders)
				bidsQty = fmt.Sprintf("%8d", bid.Quantity)
				break
			}
		}
	} else if (len(r.Asks) > 0) && (p >= r.Asks[0].Price) && (p <= r.Asks[len(r.Asks)-1].Price) {
		for _, ask := range r.Asks {
			if ask.Price == p {
				asksCount = fmt.Sprintf("%8d", ask.Orders)
				asksQty = fmt.Sprintf("%8d", ask.Quantity)
				break
			}
		}
	}

	out := fmt.Sprintf("%8d|%8.2f|%s|%s|%8d|%8d|%s|%s", pCell.VolumeTraded, p, bidsCount, bidsQty, pCell.BidQuantityTaken, pCell.AskQuantityTaken, asksQty, asksCount)
	return out

}

// StringifyTotals stringyfies total
func (r *OrderFlowMonitor) StringifyTotals() string {
	out := fmt.Sprintf("%8d                   %8d %8d %8d %8d ", r.LastTick.VolumeTraded, r.LastTick.TotalBuyQuantity, r.TotalBidsQuantityTaken, r.TotalAsksQuantityTaken, r.LastTick.TotalSellQuantity)
	return out
}
