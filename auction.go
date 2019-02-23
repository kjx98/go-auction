package auction

import (
	"errors"
	"os"

	"github.com/kjx98/avl"
	"github.com/op/go-logging"
)

type simOrderType struct {
	oid    int
	price  int
	Symbol string
	bBuy   bool
	Qty    int
}

type orderBook struct {
	bids, asks *avl.Tree
}

func (or *simOrderType) Dir() string {
	if or.bBuy {
		return "buy"
	}
	return "sell"
}

var (
	errNoOrder     = errors.New("No such order")
	errNoOrderBook = errors.New("no OrderBook")
	errCancelOrder = errors.New("can't cancel,canceled or filled")
	errOrderSeq    = errors.New("Order price disorder")
	errOrderNoSeq  = errors.New("Order same price No disorder")
)
var log = logging.MustGetLogger("go-auction")

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

// orderBook map with symbol key
var simOrderBook = map[string]orderBook{}

func cleanupOrderBook(sym string) {
	destroyTree := func(tr *avl.Tree) {
		iter := tr.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			tr.Remove(node)
		}
	}
	if orBook, ok := simOrderBook[sym]; ok {
		destroyTree(orBook.bids)
		destroyTree(orBook.asks)
		orBook.bids = nil
		orBook.asks = nil
		delete(simOrderBook, sym)
	}
}

func simInsertOrder(or *simOrderType) {
	orBook, ok := simOrderBook[or.Symbol]
	if !ok {
		orBook.bids = avl.New(bidCompare)
		orBook.asks = avl.New(askCompare)
		simOrderBook[or.Symbol] = orBook
	}
	if or.bBuy {
		// bid
		orBook.bids.Insert(or)
	} else {
		orBook.asks.Insert(or)
	}
}

func simRemoveOrder(or *simOrderType) {
	if orBook, ok := simOrderBook[or.Symbol]; ok {
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
}

func verifySimOrderBook(sym string) error {
	orB, ok := simOrderBook[sym]
	if !ok {
		log.Info("no OrderBook for ", sym)
		return errNoOrderBook
	}
	// validate bids
	iter := orB.bids.Iterator(avl.Forward)
	last := 0
	oid := 0
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		if last == 0 {
			last = v.price
			oid = v.oid
			continue
		}
		if last < v.price {
			log.Error("Bid order book price disorder for", sym)
			return errOrderSeq
		}
		if last == v.price {
			if oid > v.oid {
				log.Error("Bid order book oid disorder for", sym)
				return errOrderNoSeq
			}
			oid = v.oid
			continue
		}
		last = v.price
		oid = v.oid
	}
	// validate asks
	iter = orB.asks.Iterator(avl.Forward)
	last = 0
	oid = 0
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		if last == 0 {
			last = v.price
			oid = v.oid
			continue
		}
		if last > v.price {
			log.Error("Bid order book price disorder for", sym)
			return errOrderSeq
		}
		if last == v.price {
			if oid > v.oid {
				log.Error("Bid order book oid disorder for", sym)
				return errOrderNoSeq
			}
			oid = v.oid
			continue
		}
		last = v.price
		oid = v.oid
	}
	return nil
}

func dumpSimOrderBook(sym string) {
	orB, ok := simOrderBook[sym]
	if !ok {
		log.Info("no OrderBook for ", sym)
		return
	}
	log.Infof("Dump %s bids:", sym)
	iter := orB.bids.Iterator(avl.Forward)
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		log.Infof(" No:%d %s %d %s %d", v.oid, v.Symbol, v.price,
			v.Dir(), v.Qty)
	}
	log.Infof("Dump %s asks:", sym)
	iter = orB.asks.Iterator(avl.Forward)
	for node := iter.First(); node != nil; node = iter.Next() {
		v := node.Value.(*simOrderType)
		log.Infof(" No:%d %s %d %s %d", v.oid, v.Symbol, v.price,
			v.Dir(), v.Qty)
	}

}

func dumpSimOrderStats() {
	totalOrders := 0
	for sym, orB := range simOrderBook {
		log.Infof("%s Bid orders: %d, Ask orders: %d", sym, orB.bids.Len(), orB.asks.Len())
		totalOrders += orB.bids.Len() + orB.asks.Len()
	}
	log.Infof("Total unfilled orders: %d", totalOrders)
}

var simLogMatchs int

func simMatchOrder(ti string, isBuy bool, last, volume int) {
	setFill := func(or *simOrderType, last int, vol int) (volFilled int) {
		if vol >= or.Qty {
			volFilled = or.Qty
		} else {
			volFilled = vol
		}
		simLogMatchs++
		if simLogMatchs <= 10 {
			log.Infof("Filled No:%d %s %d %s %d(filled %d)", or.oid, or.Symbol,
				or.price, or.Dir(), or.Qty, volFilled)
		}
		return
	}

	if orB, ok := simOrderBook[ti]; ok {
		var orderQ *avl.Tree
		if isBuy {
			// fill Buy orders
			orderQ = orB.bids
		} else {
			orderQ = orB.asks
		}
		iter := orderQ.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			v := node.Value.(*simOrderType)
			if isBuy {
				if v.price >= last {
					// match
					volume -= setFill(v, last, volume)
					orderQ.Remove(node)
					if volume == 0 {
						break
					} else if volume < 0 {
						log.Error("no Way go here")
						break
					}
				} else {
					break
				}
			} else {
				if v.price <= last {
					// match
					volume -= setFill(v, last, volume)
					orderQ.Remove(node)
					if volume == 0 {
						break
					} else if volume < 0 {
						log.Error("no Way go here")
						break
					}
				} else {
					break
				}
			}
		}
	}
}

func getBestPrice(ti string, isBuy bool) int {
	if orB, ok := simOrderBook[ti]; ok {
		var orderQ *avl.Tree
		if isBuy {
			// fill Buy orders
			orderQ = orB.bids
		} else {
			orderQ = orB.asks
		}
		iter := orderQ.Iterator(avl.Forward)
		if node := iter.First(); node != nil {
			v := node.Value.(*simOrderType)
			return v.price
		}
	}
	return 0
}

func tryMatch(ti string, isBuy bool, last int) (volume, nextPrice int) {
	if orB, ok := simOrderBook[ti]; ok {
		var orderQ *avl.Tree
		if isBuy {
			// fill Buy orders
			orderQ = orB.bids
		} else {
			orderQ = orB.asks
		}
		iter := orderQ.Iterator(avl.Forward)
		for node := iter.First(); node != nil; node = iter.Next() {
			v := node.Value.(*simOrderType)
			if isBuy {
				if v.price >= last {
					// match
					volume += v.Qty
				} else {
					nextPrice = v.price
					break
				}
			} else {
				if v.price <= last {
					// match
					volume += v.Qty
				} else {
					nextPrice = v.price
					break
				}
			}
		}
	}
	return
}

func callAuction(sym string, pclose int) (last int, maxVol, volRemain int) {
	bestBid := getBestPrice(sym, true)
	bestAsk := getBestPrice(sym, false)
	if bestBid < bestAsk || bestAsk == 0 {
		return
	}
	remVol := 0
	nextP := 0
	for price := bestAsk; price != 0 && price <= bestBid; price = nextP {
		vol, _ := tryMatch(sym, true, price)
		vol2, nP := tryMatch(sym, false, price)
		nextP = nP
		if vol > vol2 {
			remVol = vol - vol2
			vol = vol2
		} else {
			remVol = vol2 - vol
		}
		if vol < maxVol {
			//continue
			break
		}
		if vol > maxVol {
			maxVol = vol
			last = price
			volRemain = remVol
		} else {
			if remVol < volRemain {
				volRemain = remVol
				last = price
			} else if remVol == volRemain {
				// last = middle last/price/pclose
				if last >= pclose {
					// unchange
				} else if pclose >= price {
					last = price
				} else {
					last = pclose
				}
			}
		}
		log.Infof("update callAuction price:%d volume:%d(left: %d)", last, maxVol, volRemain)
	}
	return
}

type quoteLevel struct {
	price  int
	volume int
}

func callAuctionNew(sym string, pclose int) (last int, maxVol, volRemain int) {
	buildQuoteLevel := func(ti string, isBuy bool, last, endPrice int) (qs []quoteLevel) {
		if orB, ok := simOrderBook[ti]; ok {
			var orderQ *avl.Tree
			if isBuy {
				// fill Buy orders
				orderQ = orB.bids
			} else {
				orderQ = orB.asks
			}
			iter := orderQ.Iterator(avl.Forward)
			volume := 0
			for node := iter.First(); node != nil; node = iter.Next() {
				v := node.Value.(*simOrderType)
				if v.price == last {
					volume += v.Qty
					continue
				}
				if volume != 0 {
					//log.Info("qs append", last, volume)
					qs = append(qs, quoteLevel{price: last, volume: volume})
				}
				volume = v.Qty
				last = v.price
				if isBuy {
					if endPrice > last {
						break
					}
				} else {
					if endPrice < last {
						break
					}
				}
			}
		}
		return
	}
	bestBid := getBestPrice(sym, true)
	bestAsk := getBestPrice(sym, false)
	if bestBid < bestAsk || bestAsk == 0 {
		return
	}
	log.Infof("BBS: %d/%d", bestBid, bestAsk)
	bidsQ := buildQuoteLevel(sym, true, bestBid, bestAsk)
	asksQ := buildQuoteLevel(sym, false, bestAsk, bestBid)
	log.Infof("bidsQ len: %d, asksQ len: %d", len(bidsQ), len(asksQ))
	//log.Info("bidsQ:", bidsQ)
	//log.Info("asksQ:", asksQ)
	i := 0 // bids
	j := 0 // asks
	bidVol := bidsQ[i].volume
	askVol := asksQ[j].volume
	bP := bidsQ[i].price
	aP := asksQ[j].price
	defer func() {
		log.Infof("update callAuctionNew price:%d volume:%d(left: %d)", last, maxVol, volRemain)
	}()
	for bP >= aP {
		switch {
		case bidVol > askVol:
			maxVol += askVol
			bidVol -= askVol
			volRemain = bidVol
			last = aP
			j++
			if j >= len(asksQ) {
				return
			}
			askVol = asksQ[j].volume
			aP = asksQ[j].price
		case bidVol < askVol:
			maxVol += bidVol
			askVol -= bidVol
			volRemain = askVol
			last = bP
			i++
			if i >= len(bidsQ) {
				return
			}
			bidVol = bidsQ[i].volume
			bP = bidsQ[i].price
		case bidVol == askVol:
			maxVol += bidVol
			volRemain = 0
			if bP == aP {
				last = bP
			} else if aP > pclose {
				last = aP
			} else if bP < pclose {
				last = bP
			} else {
				last = pclose
			}
			i++
			j++
			if i >= len(bidsQ) || j >= len(asksQ) {
				return
			}
			bidVol = bidsQ[i].volume
			bP = bidsQ[i].price
			askVol = asksQ[j].volume
			aP = asksQ[j].price
		}
		log.Infof("update callAuctionNew price:%d volume:%d(left: %d)", last, maxVol, volRemain)
	}
	log.Infof("no way go here, bp/ap: %d/%d, i/j: %d/%d", bP, aP, i, j)
	return
}

var orderNo int
var simOrders = map[int]*simOrderType{}

func SendOrder(sym string, bBuy bool, qty int, prc int) int {
	orderNo++
	var or = simOrderType{Symbol: sym, oid: orderNo, price: prc, Qty: qty, bBuy: bBuy}
	simOrders[orderNo] = &or
	// put to orderBook
	simInsertOrder(&or)
	return orderNo
}

func CancelOrder(oid int) error {
	or, ok := simOrders[oid]
	if !ok {
		return errNoOrder
	}
	simRemoveOrder(or)
	return nil
}

//  `%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`
func init() {
	var format = logging.MustStringFormatter(
		`%{color}%{time:01-02 15:04:05}  ▶ %{level:.4s} %{color:reset} %{message}`,
	)

	logback := logging.NewLogBackend(os.Stderr, "", 0)
	logfmt := logging.NewBackendFormatter(logback, format)
	logging.SetBackend(logfmt)
}
