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

var orders1 = []orderArgs{
	{"cu1906", true, 10, 42000},
	{"cu1906", true, 20, 43000},
	{"cu1906", true, 30, 41000},
	{"cu1906", true, 50, 44000},
	{"cu1906", false, 10, 45000},
	{"cu1906", false, 20, 48000},
	{"cu1906", false, 30, 46000},
	{"cu1906", false, 45, 43500},
	{"cu1906", true, 25, 43900},
	{"cu1906", false, 10, 43200},
	{"cu1906", true, 15, 43800},
	{"cu1906", false, 20, 43200},
}

var orders2 = []orderArgs{
	{"cu1906", true, 20, 43000},
	{"cu1906", true, 50, 44000},
	{"cu1906", false, 10, 45000},
	{"cu1906", false, 45, 43500},
	{"cu1906", false, 10, 43200},
	{"cu1906", true, 25, 43900},
	{"cu1906", false, 20, 43200},
}

var orders3 = []orderArgs{
	{"cu1906", true, 20, 43000},
	{"cu1906", true, 50, 44000},
	{"cu1906", true, 15, 43900},
	{"cu1906", false, 10, 45000},
	{"cu1906", false, 45, 43500},
	{"cu1906", false, 10, 43200},
	{"cu1906", false, 20, 43200},
}

var orders4 = []orderArgs{
	{"cu1906", true, 20, 43000},
	{"cu1906", true, 50, 44000},
	{"cu1906", true, 20, 43900},
	{"cu1906", true, 30, 43900},
	{"cu1906", false, 10, 45000},
	{"cu1906", false, 40, 43500},
	{"cu1906", false, 10, 43200},
	{"cu1906", false, 20, 43200},
}

var orders5 = []orderArgs{
	{"cu1906", true, 20, 43000},
	{"cu1906", true, 50, 44000},
	{"cu1906", true, 20, 43900},
	{"cu1906", false, 10, 45000},
	{"cu1906", false, 15, 43500},
	{"cu1906", false, 40, 43500},
	{"cu1906", false, 10, 43200},
	{"cu1906", false, 20, 43200},
}

var orderSS = [][]orderArgs{orders1, orders2, orders3, orders4, orders5}

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
	if err := verifySimOrderBook("cu1906"); err != nil {
		t.Error("cu1906 orderBook", err)
	}
	dumpSimOrderBook("cu1906")
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
	if err := verifySimOrderBook("cu1906"); err != nil {
		t.Error("cu1906 orderBook", err)
	}
	dumpSimOrderBook("cu1906")
	cleanupOrderBook("cu1906")
	if _, ok := simOrderBook["cu1906"]; ok {
		t.Error("cu1906 orderBook remains")
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
		{"Bid test1", args{"cu1906", true}, 44000},
		{"Ask test1", args{"cu1906", false}, 43200},
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
	tests := []struct {
		name          string
		args          args
		wantLast      int
		wantMaxVol    int
		wantVolRemain int
	}{
		// TODO: Add test cases.
		{"callAuction test1", args{"cu1906", 40000}, 43900, 75, 0},
		{"callAuction test1", args{"cu1906", 50000}, 43900, 75, 0},
	}
	cleanupOrderBook("cu1906")
	buildOrBook(orders1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := CallAuction(tt.args.sym, tt.args.pclose)
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
		{"MatchCross test1", args{"cu1906", 40000}, 1, 43900, 75, 0},
		{"MatchCross test2", args{"cu1906", 50000}, 1, 43900, 75, 0},
		{"MatchCross test3", args{"cu1906", 40000}, 2, 43500, 75, 0},
		{"MatchCross test4", args{"cu1906", 50000}, 2, 43900, 75, 0},
		{"MatchCross test5", args{"cu1906", 40000}, 3, 43900, 65, 10},
		{"MatchCross test6", args{"cu1906", 50000}, 3, 43900, 65, 10},
		{"MatchCross test7", args{"cu1906", 40000}, 4, 43500, 70, 30},
		{"MatchCross test8", args{"cu1906", 50000}, 4, 43500, 70, 30},
		{"MatchCross test9", args{"cu1906", 40000}, 5, 43900, 70, 15},
		{"MatchCross test10", args{"cu1906", 50000}, 5, 43900, 70, 15},
	}

	tdNo := 0
	//buildOrBook(orders1)
	for _, tt := range tests {
		if tt.dataNo != tdNo {
			tdNo = tt.dataNo
			if tdNo > 0 && tdNo <= len(orderSS) {
				t.Logf("Change testData orders %d", tdNo)
				cleanupOrderBook("cu1906")
				buildOrBook(orderSS[tdNo-1])
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := CallAuction(tt.args.sym, tt.args.pclose)
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
		{"MatchCross test1", args{"cu1906", 40000}, 1, 43900, 75, 0},
		{"MatchCross test2", args{"cu1906", 50000}, 1, 43900, 75, 0},
		{"MatchCross test3", args{"cu1906", 40000}, 2, 43500, 75, 0},
		{"MatchCross test4", args{"cu1906", 50000}, 2, 43900, 75, 0},
		{"MatchCross test5", args{"cu1906", 40000}, 3, 43900, 65, 10},
		{"MatchCross test6", args{"cu1906", 50000}, 3, 43900, 65, 10},
		{"MatchCross test7", args{"cu1906", 40000}, 4, 43500, 70, 30},
		{"MatchCross test8", args{"cu1906", 50000}, 4, 43500, 70, 30},
		{"MatchCross test9", args{"cu1906", 40000}, 5, 43900, 70, 15},
		{"MatchCross test10", args{"cu1906", 50000}, 5, 43900, 70, 15},
	}

	//buildOrBook(orders1)
	for _, tt := range tests {
		tdNo := tt.dataNo
		cleanupOrderBook("cu1906")
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

var pclose = 50000

func buildBenchOrderBook() {
	if ob, ok := simOrderBook["cu1908"]; ok {
		log.Infof("orderBook bids: %d, asks: %d", ob.bids.Len(), ob.asks.Len())
		return
	}
	tt := time.Now()
	rand.Seed(tt.Unix())
	//orders := []simOrderType{}
	count := int(2e6)
	for i := 0; i < count; i++ {
		price := rand.Intn(20000) + pclose - 10000
		vol := rand.Intn(100) + 1
		SendOrder("cu1908", (i&1) != 0, vol, price)
	}
	// build cu1908 orderBook
	et := time.Now()
	du := et.Sub(tt)
	log.Infof("Build rand %d orders cost %.3f seconds, %g Ops", count, du.Seconds(),
		float64(count)/du.Seconds())
	if ob, ok := simOrderBook["cu1908"]; ok {
		log.Infof("New orderBook bids: %d, asks: %d", ob.bids.Len(), ob.asks.Len())
	}
}

func BenchmarkCallAuction(b *testing.B) {
	b.StopTimer()
	buildBenchOrderBook()
	logging.SetLevel(logging.WARNING, "go-auction")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := CallAuction("cu1908", pclose)
		if i == 0 {
			b.Logf("callAuction price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}

func BenchmarkMatchCrossOld(b *testing.B) {
	b.StopTimer()
	buildBenchOrderBook()
	logging.SetLevel(logging.WARNING, "go-auction")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := MatchCrossOld("cu1908", pclose)
		if i == 0 {
			b.Logf("MatchCrossOld price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}

func BenchmarkMatchCross(b *testing.B) {
	b.StopTimer()
	buildBenchOrderBook()
	logging.SetLevel(logging.WARNING, "go-auction")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := MatchCross("cu1908", pclose)
		if i == 0 {
			b.Logf("MatchCross price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}
