package malgova

// AllocateCash book
func (b *Book) AllocateCash(Money float64) {
	b.CashAllocated = Money
	b.Cash = b.CashAllocated
}

// PlaceMarketOrder book
func (b *Book) placeMarketOrder(Qty int) {
	b.PendingOrderQuantity = Qty
	b.IsMarketOrder = true
}

// PlaceMarketOrder book
func (b *Book) placeLimitOrder(Qty int, Price float64) {
	b.PendingOrderQuantity = Qty
	b.IsMarketOrder = false
	b.PendingOrderPrice = Price
}

// QuantityAffordable book
func (b *Book) QuantityAffordable(Price float64) int {
	if Price <= b.Cash {
		return int(b.Cash / Price)
	}
	return 0
}

// Buy Order
func (b *Book) Buy(Qty int) {
	b.placeMarketOrder(Qty)
}

// Sell Order
func (b *Book) Sell(Qty int) {
	b.placeMarketOrder(-Qty)
}

// InPosition check
func (b *Book) InPosition() bool {
	return (b.Position != 0)
}

// IsOrderWaiting check
func (b *Book) IsOrderWaiting() bool {
	return (b.PendingOrderQuantity != 0)
}

// IsBookClean check
func (b *Book) IsBookClean() bool {
	return (!b.InPosition() && !b.IsOrderWaiting())
}

// Exit all position
func (b *Book) Exit() {
	if b.Position != 0 {
		b.placeMarketOrder(-b.Position)
	}
}
