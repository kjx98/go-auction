package auction

type simOrderType struct {
	oid         int
	price       int
	Symbol      string
	bBuy        bool
	Qty         int
	Filled      int
	PriceFilled int
}


func (or *simOrderType) Dir() string {
	if or.bBuy {
		return "buy"
	}
	return "sell"
}
