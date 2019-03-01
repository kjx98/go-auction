package auction

import (
	"errors"
	"os"

	"github.com/op/go-logging"
)

const maxOrders = 20000000

var orderNo int
var simOrders [maxOrders]*simOrderType

var simState int = StatePreAuction

const (
	StateIdle = iota
	StatePreAuction
	StateCallAuction
	StateTrading
	StateStop
)

var (
	errNoOrder     = errors.New("No such order")
	errNoOrderBook = errors.New("no OrderBook")
	errCancelOrder = errors.New("can't cancel,canceled or filled")
	errOrderSeq    = errors.New("Order price disorder")
	errOrderNoSeq  = errors.New("Order same price No disorder")
	errOrderFilled = errors.New("wrong order Filled volume")
	errState       = errors.New("wrong trading state")
)
var log = logging.MustGetLogger("go-auction")

// orderBook map with symbol key
var simOrderBook = map[string]*orderBook{}

func cleanupOrderBook(sym string) {
	if orBook, ok := simOrderBook[sym]; ok {
		orBook.cleanup()
		delete(simOrderBook, sym)
	}
}

func MarketStart() {
	simState = StateTrading
}

func MarketStop() {
	simState = StateStop
}

func simInsertOrder(or *simOrderType) {
	orBook, ok := simOrderBook[or.Symbol]
	if !ok {
		orBook = NewOrderBook()
		simOrderBook[or.Symbol] = orBook
	}
	orBook.insert(or)
}

func simRemoveOrder(or *simOrderType) {
	if orBook, ok := simOrderBook[or.Symbol]; ok {
		orBook.delete(or)
	}
}

func verifySimOrderBook(sym string) error {
	orB, ok := simOrderBook[sym]
	if !ok {
		log.Info("no OrderBook for ", sym)
		return errNoOrderBook
	}
	// validate bids
	last := 0
	oid := 0
	for v := orB.getBestBid(); v != nil; v = orB.nextBid() {
		if v.Filled < 0 || v.Filled > v.Qty {
			log.Errorf("Wrong Filled oid: %d Volume %d/%d", v.oid, v.Filled, v.Qty)
			return errOrderFilled
		}
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
	last = 0
	oid = 0
	for v := orB.getBestAsk(); v != nil; v = orB.nextAsk() {
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
	for v := orB.getBestBid(); v != nil; v = orB.nextBid() {
		log.Infof(" No:%d %s %d %s %d", v.oid, v.Symbol, v.price,
			v.Dir(), v.Qty)
	}
	log.Infof("Dump %s asks:", sym)
	for v := orB.getBestAsk(); v != nil; v = orB.nextAsk() {
		log.Infof(" No:%d %s %d %s %d", v.oid, v.Symbol, v.price,
			v.Dir(), v.Qty)
	}

}

func dumpSimOrderStats() {
	totalOrders := 0
	for sym, orB := range simOrderBook {
		bidLen, askLen := orB.bookLen()
		log.Infof("%s Bid orders: %d, Ask orders: %d", sym, bidLen, askLen)
		totalOrders += orB.bids.Len() + orB.asks.Len()
	}
	log.Infof("Total unfilled orders: %d", totalOrders)
}

func OrderBookLen(sym string) (bidLen, askLen int) {
	if orB, ok := simOrderBook[sym]; ok {
		bidLen, askLen = orB.bookLen()
	}
	return
}

var simLogMatchs int

func MatchOrder(sym string, isBuy bool, last, volume int) {
	setFill := func(or *simOrderType, last int, vol int) (volFilled int) {
		if vol >= or.Qty-or.Filled {
			volFilled = or.Qty - or.Filled
		} else {
			volFilled = vol
		}
		or.Filled += volFilled
		or.PriceFilled = last
		simLogMatchs++
		if simLogMatchs <= 10 {
			log.Infof("Filled No:%d %s %d %s %d(filled %d)", or.oid, or.Symbol,
				or.price, or.Dir(), or.Qty, volFilled)
		}
		return
	}

	if orB, ok := simOrderBook[sym]; ok {
		for v := orB.First(isBuy); v != nil; v = orB.Next(isBuy) {
			if isBuy {
				if v.price >= last {
					// match
					volume -= setFill(v, last, volume)
					if v.Filled >= v.Qty {
						orB.Remove(isBuy)
					} else {
						if volume > 0 {
							log.Errorf("no way go here, Buy volume @%d remains", v.price)
						}
						break
					}
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
					if v.Filled >= v.Qty {
						orB.Remove(isBuy)
					} else {
						if volume > 0 {
							log.Errorf("no way go here, Sell volume @%d remains", v.price)
						}
						break
					}
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

func tryMatchOrderBook(order *simOrderType) (filled bool) {
	setFill := func(or *simOrderType, last int, vol int) (volFilled int) {
		if vol >= or.Qty-or.Filled {
			volFilled = or.Qty - or.Filled
		} else {
			volFilled = vol
		}
		or.Filled += volFilled
		or.PriceFilled = last
		simLogMatchs++
		if simLogMatchs <= 10 {
			log.Infof("Filled No:%d %s %d %s %d(filled %d)", or.oid, or.Symbol,
				or.price, or.Dir(), or.Qty, volFilled)
		}
		return
	}

	sym := order.Symbol
	if orB, ok := simOrderBook[sym]; ok {
		isBuy := !order.bBuy
		for v := orB.First(isBuy); v != nil; v = orB.Next(isBuy) {
			volume := order.Qty - order.Filled
			last := order.price
			if isBuy {
				if v.price >= last {
					// match
					vol := setFill(v, v.price, volume)
					if v.Filled >= v.Qty {
						orB.Remove(isBuy)
					}
					setFill(order, v.price, vol)
					volume -= vol
					if volume == 0 {
						filled = true
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
					vol := setFill(v, last, volume)
					if v.Filled >= v.Qty {
						orB.Remove(isBuy)
					}
					volume -= vol
					if volume == 0 {
						filled = true
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
	return
}

func getBestPrice(ti string, isBuy bool) int {
	if orB, ok := simOrderBook[ti]; ok {
		if v := orB.First(isBuy); v != nil {
			return v.price
		}
	}
	return 0
}

func BuildOrBk(sym string) (bids, asks []*simOrderType) {
	if orB, ok := simOrderBook[sym]; ok {
		bLen, aLen := orB.bookLen()
		bids = make([]*simOrderType, bLen)
		for i, v := 0, orB.First(true); v != nil && i < len(bids); v = orB.Next(true) {
			bids[i] = v
			i++
		}
		asks = make([]*simOrderType, aLen)
		for i, v := 0, orB.First(false); v != nil && i < len(asks); v = orB.Next(false) {
			asks[i] = v
			i++
		}
	}
	return
}

func CallAuction(bids, asks []*simOrderType, pclose int) (last int, maxVol, volRemain int) {
	tryMatch := func(orders []*simOrderType, isBuy bool, last int) (volume, nextPrice int) {
		if len(orders) > 0 {
			for _, v := range orders {
				if isBuy {
					if v.price >= last {
						// match
						volume += v.Qty - v.Filled
					} else {
						nextPrice = v.price
						break
					}
				} else {
					if v.price <= last {
						// match
						volume += v.Qty - v.Filled
					} else {
						nextPrice = v.price
						break
					}
				}
			}
		}
		return
	}
	if len(bids) == 0 || len(asks) == 0 {
		return
	}
	bestBid := bids[0].price //getBestPrice(sym, true)
	bestAsk := asks[0].price //getBestPrice(sym, false)
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
			volume := 0
			for v := orB.First(isBuy); v != nil; v = orB.Next(isBuy) {
				if v.price == last {
					volume += v.Qty - v.Filled
					continue
				}
				if volume != 0 {
					qs = append(qs, quoteLevel{price: last, volume: volume})
				}
				volume = v.Qty - v.Filled
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
	var bP, aP int
	var bestBid, bestAsk int
	var bidVol, askVol int
	var orB *orderBook
	getPriceVol := func(isBuy bool) (price, vol int) {
		if v := orB.Get(isBuy); v != nil {
			price, vol = v.price, v.Qty-v.Filled
			for v = orB.Next(isBuy); v != nil; v = orB.Next(isBuy) {
				if v.price != price {
					break
				}
				vol += v.Qty - v.Filled
			}
		} else {
			price = 0
		}
		return
	}
	if orBook, ok := simOrderBook[sym]; !ok {
		return
	} else {
		orB = orBook
		if v := orB.First(true); v != nil {
			bP, bidVol = getPriceVol(true)
		}
		if v := orB.First(false); v != nil {
			aP, askVol = getPriceVol(false)
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
			aP, askVol = getPriceVol(false)
			if aP == 0 {
				break
			}
		case bidVol < askVol:
			maxVol += bidVol
			askVol -= bidVol
			volRemain = askVol
			last = bP
			bP, bidVol = getPriceVol(true)
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
			aP, askVol = getPriceVol(false)
			bP, bidVol = getPriceVol(true)
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
	var orB *orderBook
	var bidOr, askOr *simOrderType
	var bP, aP, oaP, obP int
	var bestBid, bestAsk int
	var bidVol, askVol int
	var ordersFilled = []*simOrderType{}
	getPriceVol := func(v *simOrderType) (price, vol int, or *simOrderType) {
		if v != nil {
			price, vol = v.price, v.Qty-v.Filled
			or = v
		}
		return
	}
	if orBook, ok := simOrderBook[sym]; !ok {
		return
	} else {
		orB = orBook
		bP, bidVol, bidOr = getPriceVol(orB.First(true))
		aP, askVol, askOr = getPriceVol(orB.First(false))
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
			orB.Remove(false)
			aP, askVol, askOr = getPriceVol(orB.Next(false))
		case bidVol < askVol:
			maxVol += bidVol
			askVol -= bidVol
			volRemain = askVol
			last = bP
			bidOr.Filled += askVol
			askOr.Filled += askVol
			//ordersFilled = append(ordersFilled, bidOr)
			orB.Remove(true)
			bP, bidVol, bidOr = getPriceVol(orB.Next(true))
		case bidVol == askVol:
			maxVol += bidVol
			volRemain = 0
			bidOr.Filled += askVol
			askOr.Filled += askVol
			//ordersFilled = append(ordersFilled, bidOr)
			orB.Remove(true)
			//ordersFilled = append(ordersFilled, askOr)
			orB.Remove(false)
			oaP = aP
			obP = bP
			aP, askVol, askOr = getPriceVol(orB.Next(false))
			bP, bidVol, bidOr = getPriceVol(orB.Next(true))
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
		bLen, aLen := orB.bookLen()
		log.Warningf("Last is zero!! BS QLen: %d/%d", bLen, aLen)
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
		for pr, vv, _ := getPriceVol(orB.Next(true)); pr == bP; pr, vv, _ = getPriceVol(orB.Next(true)) {
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
		for pr, vv, _ := getPriceVol(orB.Next(false)); pr == aP; pr, vv, _ = getPriceVol(orB.Next(false)) {
			volRemain += vv
		}
	}

	log.Infof("MatchCrossFill end, bp/ap: %d/%d, last/vol: %d/%d, orders filled: %d",
		bP, aP, last, maxVol, len(ordersFilled))
	return
}

func SendOrder(sym string, bBuy bool, qty int, prc int) int {
	if orderNo >= maxOrders {
		return 0
	}
	if simState == StateCallAuction || simState == StateStop {
		// wrong trading state
		return 0
	}
	var or = simOrderType{Symbol: sym, oid: orderNo + 1, price: prc, Qty: qty, bBuy: bBuy}
	simOrders[orderNo] = &or
	orderNo++
	if simState == StateTrading {
		// check match first
		if tryMatchOrderBook(&or) {
			// total filled
			return orderNo
		}
	}
	// put to orderBook
	simInsertOrder(&or)
	return orderNo
}

func CancelOrder(oid int) error {
	if simState == StateCallAuction {
		return errState
	}
	if oid <= 0 || oid > orderNo {
		return errNoOrder
	}
	or := simOrders[oid-1]
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
