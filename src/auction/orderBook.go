package auction

type orderBook struct {
	bids, asks   *Tree
	bidIt, askIt *Iterator
}

func bidCompare(a, b interface{}) int {
	ora, ok := a.(*simOrderType)
	// maybe panic, if not simOrderType
	if !ok {
		return 0
	}
	orb, ok := b.(*simOrderType)
	if !ok {
		return 0
	}
	if ora.price == orb.price {
		return ora.oid - orb.oid
	}
	if ora.price == 0 {
		return -1
	}
	if orb.price == 0 {
		return 1
	}
	// low price, low priority
	return int(orb.price) - int(ora.price)
}

func askCompare(a, b interface{}) int {
	ora, ok := a.(*simOrderType)
	// maybe panic, if not simOrderType
	if !ok {
		return 0
	}
	orb, ok := b.(*simOrderType)
	if !ok {
		return 0
	}
	if ora.price == orb.price {
		return ora.oid - orb.oid
	}
	// high price, low priority
	return int(ora.price) - int(orb.price)
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

func (orB *orderBook) Remove(isBuy bool) {
	if isBuy {
		orB.bidIt.RemoveCur()
	} else {
		orB.askIt.RemoveCur()
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
		or := v.(*simOrderType)
		return or
	}
	return nil
}

func (orB *orderBook) nextBid() *simOrderType {
	if orB.bidIt != nil {
		if v := orB.bidIt.Next(); v != nil {
			or := v.(*simOrderType)
			return or
		}
	}
	return nil
}

func (orB *orderBook) curBid() *simOrderType {
	if orB.bidIt == nil {
		return nil
	}
	if v := orB.bidIt.Get(); v != nil {
		or := v.(*simOrderType)
		return or
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
		or := v.(*simOrderType)
		return or
	}
	return nil
}

func (orB *orderBook) nextAsk() *simOrderType {
	if orB.askIt != nil {
		if v := orB.askIt.Next(); v != nil {
			or := v.(*simOrderType)
			return or
		}
	}
	return nil
}

func (orB *orderBook) curAsk() *simOrderType {
	if orB.askIt == nil {
		return nil
	}
	if v := orB.askIt.Get(); v != nil {
		or := v.(*simOrderType)
		return or
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
