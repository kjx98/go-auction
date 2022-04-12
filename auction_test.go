package auction

import (
	"math/rand"
	"testing"
	"time"

	logging "github.com/op/go-logging"
)

type orderArgs struct {
	sym  string
	bBuy bool
	qty  int
	prc  int
}

type dealArgs struct {
	no    int
	oid   int
	price int
	vol   int
}

var testInstr = "cu1906"
var orders1 = []orderArgs{
	{testInstr, true, 10, 42000},
	{testInstr, true, 20, 43000},
	{testInstr, true, 30, 41000},
	{testInstr, true, 50, 44000},
	{testInstr, false, 10, 45000},
	{testInstr, false, 20, 48000},
	{testInstr, false, 30, 46000},
	{testInstr, false, 45, 43500},
	{testInstr, true, 25, 43900},
	{testInstr, false, 10, 43200},
	{testInstr, true, 15, 43800},
	{testInstr, false, 20, 43200},
}

var deals1 = []dealArgs{
	{1, 4, 43500, 45},
	{2, 8, 43500, 45},
	{3, 4, 43200, 5},
	{4, 10, 43200, 5},
	{5, 9, 43200, 5},
	{6, 10, 43200, 5},
	{7, 9, 43200, 20},
	{8, 12, 43200, 20},
}

var orders2 = []orderArgs{
	{testInstr, true, 20, 43000},
	{testInstr, true, 50, 44000},
	{testInstr, false, 10, 45000},
	{testInstr, false, 45, 43500},
	{testInstr, false, 10, 43200},
	{testInstr, true, 25, 43900},
	{testInstr, false, 20, 43200},
}

var orders3 = []orderArgs{
	{testInstr, true, 20, 43000},
	{testInstr, true, 50, 44000},
	{testInstr, true, 15, 43900},
	{testInstr, false, 10, 45000},
	{testInstr, false, 45, 43500},
	{testInstr, false, 10, 43200},
	{testInstr, false, 20, 43200},
}

var orders4 = []orderArgs{
	{testInstr, true, 20, 43000},
	{testInstr, true, 50, 44000},
	{testInstr, true, 20, 43900},
	{testInstr, true, 30, 43900},
	{testInstr, false, 10, 45000},
	{testInstr, false, 40, 43500},
	{testInstr, false, 10, 43200},
	{testInstr, false, 20, 43200},
}

var orders5 = []orderArgs{
	{testInstr, true, 20, 43000},
	{testInstr, true, 50, 44000},
	{testInstr, true, 20, 43900},
	{testInstr, false, 10, 45000},
	{testInstr, false, 15, 43500},
	{testInstr, false, 40, 43500},
	{testInstr, false, 10, 43200},
	{testInstr, false, 20, 43200},
}

var orderSS = [][]orderArgs{orders1, orders2, orders3, orders4, orders5}

func init() {
	logging.SetLevel(logging.WARNING, "go-auction")
}

func buildOrBook(orders []orderArgs) {
	for _, or := range orders {
		if nn := SendOrder(or.sym, or.bBuy, or.qty, or.prc); nn == 0 {
			log.Errorf("SendOrder price:%d vol:%d, No: %d", or.prc, or.qty, nn)
		}
	}
}

func TestSendOrder(t *testing.T) {
	tests := []struct {
		name string
		args orderArgs
		want int
	}{
		// TODO: Add test cases.
		{"cu1906-t1", orders1[0], 1},
		{"cu1906-t2", orders1[1], 2},
		{"cu1906-t3", orders1[2], 3},
		{"cu1906-t4", orders1[3], 4},
		{"cu1906-t5", orders1[4], 5},
		{"cu1906-t6", orders1[5], 6},
		{"cu1906-t7", orders1[6], 7},
		{"cu1906-t8", orders1[7], 8},
		{"cu1906-t9", orders1[8], 9},
		{"cu1906-t10", orders1[9], 10},
		{"cu1906-t11", orders1[10], 11},
		{"cu1906-t12", orders1[11], 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SendOrder(tt.args.sym, tt.args.bBuy, tt.args.qty, tt.args.prc); got != tt.want {
				t.Errorf("SendOrder() = %v, want %v", got, tt.want)
			}
		})
	}
	if err := verifySimOrderBook(testInstr); err != nil {
		t.Error("cu1906 orderBook", err)
	}
	dumpSimOrderBook(testInstr)
}

func TestCancelOrder(t *testing.T) {
	type args struct {
		oid int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"cancelOrder1", args{3}, false},
		{"cancelOrder2", args{7}, false},
		{"cancelOrder3", args{23}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CancelOrder(tt.args.oid); (err != nil) != tt.wantErr {
				t.Errorf("CancelOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := verifySimOrderBook(testInstr); err != nil {
		t.Error(testInstr, "orderBook", err)
	}
	dumpSimOrderBook(testInstr)
	cleanupOrderBook(testInstr)
	if _, ok := simOrderBook[testInstr]; ok {
		t.Error(testInstr, "orderBook remains")
	}
}

func Test_getBestPrice(t *testing.T) {
	type args struct {
		ti    string
		isBuy bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{"Bid test1", args{testInstr, true}, 44000},
		{"Ask test1", args{testInstr, false}, 43200},
	}
	buildOrBook(orders1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBestPrice(tt.args.ti, tt.args.isBuy); got != tt.want {
				t.Errorf("getBestPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallAuction(t *testing.T) {
	type args struct {
		sym    string
		pclose int
	}
	instr := testInstr
	tests := []struct {
		name          string
		args          args
		wantLast      int
		wantMaxVol    int
		wantVolRemain int
	}{
		// TODO: Add test cases.
		{"callAuction test1", args{instr, 40000}, 43900, 75, 0},
		{"callAuction test1", args{instr, 50000}, 43900, 75, 0},
	}
	cleanupOrderBook(instr)
	buildOrBook(orders1)
	bids, asks := BuildOrBk(instr)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := CallAuction(bids, asks, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("callAuction() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("callAuction() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("callAuction() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
			}
		})
	}
}

func TestMatchCross(t *testing.T) {
	type args struct {
		sym    string
		pclose int
	}
	tests := []struct {
		name          string
		args          args
		dataNo        int
		wantLast      int
		wantMaxVol    int
		wantVolRemain int
	}{
		// TODO: Add test cases.
		{"MatchCross test1", args{testInstr, 40000}, 1, 43900, 75, 0},
		{"MatchCross test2", args{testInstr, 50000}, 1, 43900, 75, 0},
		{"MatchCross test3", args{testInstr, 40000}, 2, 43500, 75, 0},
		{"MatchCross test4", args{testInstr, 50000}, 2, 43900, 75, 0},
		{"MatchCross test5", args{testInstr, 40000}, 3, 43900, 65, 10},
		{"MatchCross test6", args{testInstr, 50000}, 3, 43900, 65, 10},
		{"MatchCross test7", args{testInstr, 40000}, 4, 43500, 70, 30},
		{"MatchCross test8", args{testInstr, 50000}, 4, 43500, 70, 30},
		{"MatchCross test9", args{testInstr, 40000}, 5, 43900, 70, 15},
		{"MatchCross test10", args{testInstr, 50000}, 5, 43900, 70, 15},
	}

	tdNo := 0
	var bids, asks []*simOrderType
	//buildOrBook(orders1)
	for _, tt := range tests {
		if tt.dataNo != tdNo {
			tdNo = tt.dataNo
			if tdNo > 0 && tdNo <= len(orderSS) {
				t.Logf("Change testData orders %d", tdNo)
				cleanupOrderBook(testInstr)
				buildOrBook(orderSS[tdNo-1])
				bids, asks = BuildOrBk(tt.args.sym)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := CallAuction(bids, asks, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("callAuction() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("callAuction() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("callAuction() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
			}
			gotLast, gotMaxVol, gotVolRemain = MatchCrossOld(tt.args.sym, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("MatchCrossOld() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("MatchCrossOld() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("MatchCrossOld() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
			}
			gotLast, gotMaxVol, gotVolRemain = MatchCross(tt.args.sym, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("MatchCross() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("MatchCross() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("MatchCross() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
			}
		})
	}
}

func TestMatchCrossFill(t *testing.T) {
	type args struct {
		sym    string
		pclose int
	}
	tests := []struct {
		name          string
		args          args
		dataNo        int
		wantLast      int
		wantMaxVol    int
		wantVolRemain int
	}{
		// TODO: Add test cases.
		{"MatchCross test1", args{testInstr, 40000}, 1, 43900, 75, 0},
		{"MatchCross test2", args{testInstr, 50000}, 1, 43900, 75, 0},
		{"MatchCross test3", args{testInstr, 40000}, 2, 43500, 75, 0},
		{"MatchCross test4", args{testInstr, 50000}, 2, 43900, 75, 0},
		{"MatchCross test5", args{testInstr, 40000}, 3, 43900, 65, 10},
		{"MatchCross test6", args{testInstr, 50000}, 3, 43900, 65, 10},
		{"MatchCross test7", args{testInstr, 40000}, 4, 43500, 70, 30},
		{"MatchCross test8", args{testInstr, 50000}, 4, 43500, 70, 30},
		{"MatchCross test9", args{testInstr, 40000}, 5, 43900, 70, 15},
		{"MatchCross test10", args{testInstr, 50000}, 5, 43900, 70, 15},
	}

	//buildOrBook(orders1)
	for _, tt := range tests {
		tdNo := tt.dataNo
		cleanupOrderBook(testInstr)
		buildOrBook(orderSS[tdNo-1])
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := MatchCrossFill(tt.args.sym, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("MatchCrossFill() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("MatchCrossFill() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("MatchCrossFill() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
			}
		})
	}
}

func TestTraingContinue(t *testing.T) {
	tests := []struct {
		name string
		args orderArgs
		want int
	}{
		// TODO: Add test cases.
		{"cu1906-t1", orders1[0], 1},
		{"cu1906-t2", orders1[1], 2},
		{"cu1906-t3", orders1[2], 3},
		{"cu1906-t4", orders1[3], 4},
		{"cu1906-t5", orders1[4], 5},
		{"cu1906-t6", orders1[5], 6},
		{"cu1906-t7", orders1[6], 7},
		{"cu1906-t8", orders1[7], 8},
		{"cu1906-t9", orders1[8], 9},
		{"cu1906-t10", orders1[9], 10},
		{"cu1906-t11", orders1[10], 11},
		{"cu1906-t12", orders1[11], 12},
	}
	cleanupOrderBook(testInstr)
	MarketStart(true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SendOrder(tt.args.sym, tt.args.bBuy, tt.args.qty, tt.args.prc); got != tt.want {
				t.Errorf("SendOrder() = %v, want %v", got, tt.want)
			}
		})
	}
	if err := verifySimOrderBook(testInstr); err != nil {
		t.Error("cu1906 orderBook", err)
	}
	MarketStop()
	// verify Deals
	for _, dd := range deals1 {
		if dv := getDeal(dd.no); dv == nil {
			t.Errorf("Can't find DealNo: %d", dd.no)
		} else {
			if dd.oid != dv.oid {
				t.Errorf("DealNo: %d, oid differ: %d Got %d", dd.no, dd.oid, dv.oid)
			}
			if dd.price != dv.price {
				t.Errorf("DealNo: %d, price differ: %d Got %d", dd.no, dd.price, dv.price)
			}
			if dd.vol != dv.vol {
				t.Errorf("DealNo: %d, volume differ: %d Got %d", dd.no, dd.vol, dv.vol)
			}
		}
	}
	dumpSimOrderBook(testInstr)
}

var pclose = 50000

func buildBenchOrderBook(instr string) int {
	if ob, ok := simOrderBook[instr]; ok {
		bLen, aLen := ob.bookLen()
		log.Infof("orderBook bids: %d, asks: %d", bLen, aLen)
		return bLen + aLen
	}
	simState = StatePreAuction
	tt := time.Now()
	rand.Seed(tt.Unix())
	//orders := []simOrderType{}
	count := int(2e6)
	for i := 0; i < count; i++ {
		price := rand.Intn(20000) + pclose - 10000
		vol := rand.Intn(100) + 1
		SendOrder(instr, (i&1) != 0, vol, price)
	}
	// build cu1908 orderBook
	et := time.Now()
	du := et.Sub(tt)
	log.Infof("Build rand %d orders cost %.3f seconds, %g Ops", count, du.Seconds(),
		float64(count)/du.Seconds())
	if ob, ok := simOrderBook[instr]; ok {
		bLen, aLen := ob.bookLen()
		log.Infof("New orderBook bids: %d, asks: %d", bLen, aLen)
	}
	return count
}

func BenchmarkCallAuction(b *testing.B) {
	b.StopTimer()
	instr := "cu1908"
	buildBenchOrderBook(instr)
	logging.SetLevel(logging.WARNING, "go-auction")
	bids, asks := BuildOrBk(instr)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := CallAuction(bids, asks, pclose)
		if i == 0 {
			b.Logf("callAuction price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}

func BenchmarkMatchCrossOld(b *testing.B) {
	b.StopTimer()
	instr := "cu1908"
	buildBenchOrderBook(instr)
	logging.SetLevel(logging.WARNING, "go-auction")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		MatchCrossOld(instr, pclose)
		/*
			gotLast, vol, gotVolRemain := MatchCrossOld(instr, pclose)
			if i == 0 {
				b.Logf("MatchCrossOld price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
			}
		*/
	}
}

func BenchmarkMatchCross(b *testing.B) {
	b.StopTimer()
	instr := "cu1908"
	buildBenchOrderBook(instr)
	logging.SetLevel(logging.WARNING, "go-auction")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		MatchCross(instr, pclose)
		/*
			gotLast, vol, gotVolRemain := MatchCross(instr, pclose)
				if i == 0 {
					b.Logf("MatchCross price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
				}
		*/
	}
}

func BenchmarkMatchTradeContinue(b *testing.B) {
	b.StopTimer()
	instr := "cu1908"
	buildBenchOrderBook(instr)
	logging.SetLevel(logging.WARNING, "go-auction")
	//last, vol, volRemain := MatchCross(instr, pclose)
	last, vol, _ := MatchCross(instr, pclose)
	//b.Logf("MatchCross price:%d, volume:%d, remainVol:%d", last, vol, volRemain)
	if last > 0 {
		MatchOrder(instr, true, last, vol)
		MatchOrder(instr, false, last, vol)
	}
	/*
		if ob, ok := simOrderBook[instr]; ok {
			bLen, aLen := ob.bookLen()
			b.Logf("Before TradeContinous orderBook bids: %d, asks: %d", bLen, aLen)
		}
	*/
	logging.SetLevel(logging.WARNING, "go-auction")
	MarketStart(false)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		price := rand.Intn(20000) + pclose - 10000
		vol := rand.Intn(100) + 1
		SendOrder(instr, (price&1) != 0, vol, price)
	}
	b.StopTimer()
	MarketStop()
	/*
		if ob, ok := simOrderBook[instr]; ok {
			bLen, aLen := ob.bookLen()
			b.Logf("After TradeContinous orderBook bids: %d, asks: %d", bLen, aLen)
		}
	*/
}
