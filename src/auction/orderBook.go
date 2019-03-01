package auction

import (
	"github.com/kjx98/avl"
)

type orderBook struct {
	bids, asks   *avl.Tree
	bidIt, askIt *avl.Iterator
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
	destroyTree := func(tr *avl.Tree) {
		iter := tr.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			tr.Remove(node)
		}
	}
	destroyTree(orBook.bids)
	destroyTree(orBook.asks)
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
		if v := orBook.bids.Find(or); v != nil {
			orBook.bids.Remove(v)
		}
	} else {
		if v := orBook.asks.Find(or); v != nil {
			orBook.asks.Remove(v)
		}
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
		node := orB.bidIt.Get()
		orB.bids.Remove(node)
	} else {
		node := orB.askIt.Get()
		orB.asks.Remove(node)
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
		orB.bidIt = orB.bids.Iterator(avl.Forward)
	}
	if node := orB.bidIt.First(); node != nil {
		or := node.Value.(*simOrderType)
		return or
	}
	return nil
}

func (orB *orderBook) nextBid() *simOrderType {
	if orB.bidIt != nil {
		if node := orB.bidIt.Next(); node != nil {
			v := node.Value.(*simOrderType)
			return v
		}
	}
	return nil
}

func (orB *orderBook) curBid() *simOrderType {
	if orB.bidIt == nil {
		return nil
	}
	if node := orB.bidIt.Get(); node != nil {
		or := node.Value.(*simOrderType)
		return or
	}
	return nil
}

func (orB *orderBook) getBestAsk() *simOrderType {
	if orB.askIt == nil {
		orB.askIt = orB.asks.Iterator(avl.Forward)
	}
	if node := orB.askIt.First(); node != nil {
		or := node.Value.(*simOrderType)
		return or
	}
	return nil
}

func (orB *orderBook) nextAsk() *simOrderType {
	if orB.askIt != nil {
		if node := orB.askIt.Next(); node != nil {
			or := node.Value.(*simOrderType)
			return or
		}
	}
	return nil
}

func (orB *orderBook) curAsk() *simOrderType {
	if orB.askIt == nil {
		return nil
	}
	if node := orB.askIt.Get(); node != nil {
		or := node.Value.(*simOrderType)
		return or
	}
	return nil
}

func NewOrderBook() *orderBook {
	var orBook orderBook
	orBook.bids = avl.New(bidCompare)
	orBook.asks = avl.New(askCompare)
	orBook.bidIt = nil
	orBook.askIt = nil
	return &orBook
}
