package auction

import (
	"errors"
	"os"

	"github.com/kjx98/avl"
	"github.com/op/go-logging"
)

type simOrderType struct {
	oid         int
	price       int
	Symbol      string
	bBuy        bool
	Qty         int
	Filled      int
	PriceFilled int
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

func OrderBookLen(sym string) (bidLen, askLen int) {
	if orB, ok := simOrderBook[sym]; ok {
		bidLen, askLen = orB.bids.Len(), orB.asks.Len()
	}
	return
}

var simLogMatchs int

func MatchOrder(sym string, isBuy bool, last, volume int) {
	setFill := func(or *simOrderType, last int, vol int) (volFilled int) {
		if vol >= or.Qty {
			volFilled = or.Qty
		} else {
			volFilled = vol
		}
		or.Filled = volFilled
		or.PriceFilled = last
		simLogMatchs++
		if simLogMatchs <= 10 {
			log.Infof("Filled No:%d %s %d %s %d(filled %d)", or.oid, or.Symbol,
				or.price, or.Dir(), or.Qty, volFilled)
		}
		return
	}

	if orB, ok := simOrderBook[sym]; ok {
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

func buildOrBk(sym string) (bids, asks []*simOrderType) {
	if orB, ok := simOrderBook[sym]; ok {
		bids = make([]*simOrderType, orB.bids.Len())
		it := orB.bids.Iterator(avl.Forward)
		for i, node := 0, it.First(); node != nil && i < len(bids); node = it.Next() {
			v := node.Value.(*simOrderType)
			bids[i] = v
			i++
		}
		asks = make([]*simOrderType, orB.asks.Len())
		it = orB.asks.Iterator(avl.Forward)
		for i, node := 0, it.First(); node != nil && i < len(asks); node = it.Next() {
			v := node.Value.(*simOrderType)
			asks[i] = v
			i++
		}
	}
	return
}

func CallAuction(sym string, pclose int) (last int, maxVol, volRemain int) {
	tryMatch := func(orders []*simOrderType, isBuy bool, last int) (volume, nextPrice int) {
		if len(orders) > 0 {
			for _, v := range orders {
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
	bids, asks := buildOrBk(sym)
	bestBid := getBestPrice(sym, true)
	bestAsk := getBestPrice(sym, false)
	if bestBid < bestAsk || bestAsk == 0 {
		return
	}
	log.Infof("callAuction BBS: %d/%d, BS QLen: %d/%d", bestBid, bestAsk, len(bids), len(asks))
	remVol := 0
	nextP := 0
	vol := 0
	for price := bestAsk; price != 0 && price <= bestBid; price = nextP {
		bVol, _ := tryMatch(bids, true, price)
		aVol, aP := tryMatch(asks, false, price)
		nextP = aP
		if bVol > aVol {
			remVol = bVol - aVol
			vol = aVol
		} else {
			remVol = aVol - bVol
			vol = bVol
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
			log.Info("no way go here, scanup cant duplicate maxVol")
			break
		}
		log.Infof("update callAuction price:%d volume:%d(left: %d) nextP:%d", last, maxVol, volRemain, nextP)
	}
	// scan down
	tmpLast := last
	log.Infof("scan down %d to %d", bestBid, tmpLast)
	for price := bestBid; price != 0 && price > tmpLast; price = nextP {
		bVol, bP := tryMatch(bids, true, price)
		aVol, _ := tryMatch(asks, false, price)
		nextP = bP
		if bVol > aVol {
			remVol = bVol - aVol
			vol = aVol
		} else {
			remVol = aVol - bVol
			vol = bVol
		}
		if vol < maxVol {
			continue
		}
		if vol > maxVol {
			maxVol = vol
			last = price
			volRemain = remVol
			log.Info("no way go here, scandown new maxVol")
		} else {
			if remVol < volRemain {
				volRemain = remVol
				last = price
			} else if remVol == volRemain {
				if bVol < aVol {
					last = price
				} else if bVol == aVol {
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
		}
		log.Infof("update callAuction (down) price:%d volume:%d(left: %d) nextP:%d", last, maxVol, volRemain, nextP)
		//break
	}
	log.Infof("callAuction end price:%d volume:%d(left: %d)", last, maxVol, volRemain)
	return
}

type quoteLevel struct {
	price  int
	volume int
}

func MatchCrossOld(sym string, pclose int) (last int, maxVol, volRemain int) {
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
	log.Infof("MatchCrossOld BBS: %d/%d", bestBid, bestAsk)
	bidsQ := buildQuoteLevel(sym, true, bestBid, bestAsk)
	asksQ := buildQuoteLevel(sym, false, bestAsk, bestBid)
	log.Infof("bidsQ len: %d, asksQ len: %d", len(bidsQ), len(asksQ))
	i := 0 // bids
	j := 0 // asks
	bidVol := bidsQ[i].volume
	askVol := asksQ[j].volume
	bP := bidsQ[i].price
	aP := asksQ[j].price
	for aP != 0 && bP >= aP {
		switch {
		case bidVol > askVol:
			maxVol += askVol
			bidVol -= askVol
			volRemain = bidVol
			last = aP
			j++
			if j >= len(asksQ) {
				aP = 0
				break
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
				bP = 0
				break
			}
			bidVol = bidsQ[i].volume
			bP = bidsQ[i].price
		case bidVol == askVol:
			maxVol += bidVol
			volRemain = 0
			if bP == aP {
				last = bP
				break
			}
			i++
			j++
			if i >= len(bidsQ) && j >= len(asksQ) {
				if aP > pclose {
					last = aP
				} else if bP < pclose {
					last = bP
				} else {
					last = pclose
				}
				aP = 0
				bP = 0
				break
			}
			if i >= len(bidsQ) {
				last = aP
				bP = 0
				break
			}
			if j >= len(asksQ) {
				last = bP
				aP = 0
				break
			}
			bidVol = bidsQ[i].volume
			bP = bidsQ[i].price
			askVol = asksQ[j].volume
			aP = asksQ[j].price
			if bidVol <= askVol {
				last = bP
			} else {
				last = aP
			}
		}
		log.Infof("update MatchCrossOld price:%d volume:%d(left: %d)", last, maxVol, volRemain)
	}
	log.Infof("MatchCrossOld end, bp/ap: %d/%d, i/j: %d/%d", bP, aP, i, j)
	return
}

func MatchCross(sym string, pclose int) (last int, maxVol, volRemain int) {
	var bidIter, askIter *avl.Iterator
	var bP, aP int
	var bestBid, bestAsk int
	var bidVol, askVol int
	getPriceVol := func(it *avl.Iterator) (price, vol int) {
		if node := it.Get(); node != nil {
			v := node.Value.(*simOrderType)
			price, vol = v.price, v.Qty
			for node = it.Next(); node != nil; node = it.Next() {
				v = node.Value.(*simOrderType)
				if v.price != price {
					break
				}
				vol += v.Qty
			}
		} else {
			price = 0
		}
		return
	}
	if orB, ok := simOrderBook[sym]; !ok {
		return
	} else {
		bidIter = orB.bids.Iterator(avl.Forward)
		if node := bidIter.First(); node != nil {
			bP, bidVol = getPriceVol(bidIter)
		}
		askIter = orB.asks.Iterator(avl.Forward)
		if node := askIter.First(); node != nil {
			aP, askVol = getPriceVol(askIter)
		}
	}

	if bP < aP || aP == 0 {
		return
	}
	bestBid, bestAsk = bP, aP
	log.Infof("MatchCross BBS: %d/%d", bP, aP)

	for aP != 0 && bP >= aP {
		switch {
		case bidVol > askVol:
			maxVol += askVol
			bidVol -= askVol
			volRemain = bidVol
			last = aP
			aP, askVol = getPriceVol(askIter)
			if aP == 0 {
				break
			}
		case bidVol < askVol:
			maxVol += bidVol
			askVol -= bidVol
			volRemain = askVol
			last = bP
			bP, bidVol = getPriceVol(bidIter)
			if bP == 0 {
				break
			}
		case bidVol == askVol:
			maxVol += bidVol
			volRemain = 0
			if bP == aP {
				// should be other bids/asks
				last = bP
				break
			}
			oaP := aP
			obP := bP
			aP, askVol = getPriceVol(askIter)
			bP, bidVol = getPriceVol(bidIter)
			if aP > bestBid {
				aP = 0
			}
			if bP < bestAsk {
				bP = 0
			}
			if bP == 0 && aP == 0 {
				if oaP > pclose {
					last = oaP
				} else if obP < pclose {
					last = obP
				} else {
					last = pclose
				}
				break
			}
			if bP == 0 {
				last = oaP
			}
			if aP == 0 {
				last = obP
			}
		}
		log.Infof("update MatchCross price:%d %d/%d volume:%d(left: %d)", last, bP, aP, maxVol, volRemain)
	}
	log.Infof("MatchCross end, bp/ap: %d/%d", bP, aP)
	return
}

func MatchCrossFill(sym string, pclose int) (last int, maxVol, volRemain int) {
	var bidIter, askIter *avl.Iterator
	var bidsTree, asksTree *avl.Tree
	var bidOr, askOr *simOrderType
	var bP, aP, oaP, obP int
	var bestBid, bestAsk int
	var bidVol, askVol int
	var ordersFilled = []*simOrderType{}
	getPriceVol := func(node *avl.Node) (price, vol int, or *simOrderType) {
		if node != nil {
			v := node.Value.(*simOrderType)
			price, vol = v.price, v.Qty
			or = v
		}
		return
	}
	if orB, ok := simOrderBook[sym]; !ok {
		return
	} else {
		bidsTree, asksTree = orB.bids, orB.asks
		bidIter = orB.bids.Iterator(avl.Forward)
		bP, bidVol, bidOr = getPriceVol(bidIter.First())
		askIter = orB.asks.Iterator(avl.Forward)
		aP, askVol, askOr = getPriceVol(askIter.First())
	}

	if bP < aP || aP == 0 {
		return
	}
	bestBid, bestAsk = bP, aP
	log.Infof("MatchCrossFill BBS: %d/%d", bP, aP)

	for aP != 0 && bP >= aP {
		switch {
		case bidVol > askVol:
			maxVol += askVol
			bidVol -= askVol
			volRemain = bidVol
			last = aP
			bidOr.Filled += askVol
			askOr.Filled += askVol
			//ordersFilled = append(ordersFilled, askOr)
			asksTree.Remove(askIter.Get())
			aP, askVol, askOr = getPriceVol(askIter.Next())
		case bidVol < askVol:
			maxVol += bidVol
			askVol -= bidVol
			volRemain = askVol
			last = bP
			bidOr.Filled += askVol
			askOr.Filled += askVol
			//ordersFilled = append(ordersFilled, bidOr)
			bidsTree.Remove(bidIter.Get())
			bP, bidVol, bidOr = getPriceVol(bidIter.Next())
		case bidVol == askVol:
			maxVol += bidVol
			volRemain = 0
			bidOr.Filled += askVol
			askOr.Filled += askVol
			//ordersFilled = append(ordersFilled, bidOr)
			bidsTree.Remove(bidIter.Get())
			//ordersFilled = append(ordersFilled, askOr)
			asksTree.Remove(askIter.Get())
			oaP = aP
			obP = bP
			aP, askVol, askOr = getPriceVol(askIter.Next())
			bP, bidVol, bidOr = getPriceVol(bidIter.Next())
			if obP == oaP {
				// maybe other bids or asks left
				last = obP
				break
			}
			if aP > bestBid {
				aP = 0
			}
			if bP < bestAsk {
				bP = 0
			}
			if bP == 0 && aP == 0 {
				if oaP > pclose {
					last = oaP
				} else if obP < pclose {
					last = obP
				} else {
					last = pclose
				}
				break
			}
			if bP == 0 {
				if aP == oaP {
					last = obP
				} else {
					last = oaP
				}
			}
			if aP == 0 {
				if bP == obP {
					last = oaP
				} else {
					last = obP
				}
			}
		}
		log.Infof("update MatchCrossFill price:%d %d/%d volume:%d(left: %d)", last, bP, aP, maxVol, volRemain)
	}
	if last == 0 {
		log.Warningf("Last is zero!! BS QLen: %d/%d", bidsTree.Len(), asksTree.Len())
		return
	}
	// fix volRemain
	switch {
	case oaP == last && volRemain == 0:
		// fix pr
		if bP == 0 {
			break
		}
		volRemain = bidVol
		fallthrough
	case bP == last && volRemain != 0:
		log.Infof("fix bidVol, volRemain, last/price: %d/%d", last, bP)
		for pr, vv, _ := getPriceVol(bidIter.Next()); pr == bP; pr, vv, _ = getPriceVol(bidIter.Next()) {
			volRemain += vv
		}
	case obP == last && volRemain == 0:
		// fix pr
		if aP == 0 {
			break
		}
		volRemain = askVol
		fallthrough
	case aP == last && volRemain != 0:
		log.Infof("fix askVol, volRemain, last/price: %d/%d", last, aP)
		for pr, vv, _ := getPriceVol(askIter.Next()); pr == aP; pr, vv, _ = getPriceVol(askIter.Next()) {
			volRemain += vv
		}
	}

	log.Infof("MatchCrossFill end, bp/ap: %d/%d, last/vol: %d/%d, orders filled: %d",
		bP, aP, last, maxVol, len(ordersFilled))
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
