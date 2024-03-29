package auction

type orderBook struct {
	bids, asks   *Tree
	bidIt, askIt *Iterator
}

func bidCompare(a, b *simOrderType) int {
	if a.price == b.price {
		return a.oid - b.oid
	}
	if a.price == 0 {
		return -1
	}
	if b.price == 0 {
		return 1
	}
	// low price, low priority
	return int(b.price) - int(a.price)
}

func askCompare(a, b *simOrderType) int {
	if a.price == b.price {
		return a.oid - b.oid
	}
	// high price, low priority
	return int(a.price) - int(b.price)
}

func (orBook *orderBook) cleanup() {
	orBook.bids.destroy()
	orBook.asks.destroy()
	orBook.bids = nil
	orBook.asks = nil
}

func (orB *orderBook) bookLen() (int, int) {
	return orB.bids.Len(), orB.asks.Len()
}

func (orBook *orderBook) insert(or *simOrderType) {
	if or.bBuy {
		// bid
		orBook.bids.Insert(or)
	} else {
		orBook.asks.Insert(or)
	}
}

func (orBook *orderBook) delete(or *simOrderType) {
	if or.bBuy {
		orBook.bids.Delete(or)
	} else {
		orBook.asks.Delete(or)
	}
}

func (orB *orderBook) First(isBuy bool) *simOrderType {
	if isBuy {
		return orB.getBestBid()
	}
	return orB.getBestAsk()
}

func (orB *orderBook) RemoveFirst(isBuy bool) {
	if isBuy {
		orB.bidIt.RemoveFirst()
	} else {
		orB.askIt.RemoveFirst()
	}

}

func (orB *orderBook) Next(isBuy bool) *simOrderType {
	if isBuy {
		return orB.nextBid()
	}
	return orB.nextAsk()
}

func (orB *orderBook) Get(isBuy bool) *simOrderType {
	if isBuy {
		return orB.curBid()
	}
	return orB.curAsk()
}

func (orB *orderBook) getBestBid() *simOrderType {
	if orB.bidIt == nil {
		orB.bidIt = orB.bids.First()
	} else {
		orB.bidIt.First()
	}
	if v := orB.bidIt.Get(); v != nil {
		return v
	}
	return nil
}

func (orB *orderBook) nextBid() *simOrderType {
	if orB.bidIt != nil {
		if v := orB.bidIt.Next(); v != nil {
			return v
		}
	}
	return nil
}

func (orB *orderBook) curBid() *simOrderType {
	if orB.bidIt == nil {
		return nil
	}
	if v := orB.bidIt.Get(); v != nil {
		return v
	}
	return nil
}

func (orB *orderBook) getBestAsk() *simOrderType {
	if orB.askIt == nil {
		orB.askIt = orB.asks.First()
	} else {
		orB.askIt.First()
	}
	if v := orB.askIt.Get(); v != nil {
		return v
	}
	return nil
}

func (orB *orderBook) nextAsk() *simOrderType {
	if orB.askIt != nil {
		if v := orB.askIt.Next(); v != nil {
			return v
		}
	}
	return nil
}

func (orB *orderBook) curAsk() *simOrderType {
	if orB.askIt == nil {
		return nil
	}
	if v := orB.askIt.Get(); v != nil {
		return v
	}
	return nil
}

func NewOrderBook() *orderBook {
	var orBook orderBook
	orBook.bids = NewTree(bidCompare)
	orBook.asks = NewTree(askCompare)
	orBook.bidIt = nil
	orBook.askIt = nil
	return &orBook
}
