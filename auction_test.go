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
	{"cu1906", true, 25, 43800},
	{"cu1906", false, 10, 43200},
	{"cu1906", true, 15, 43800},
	{"cu1906", false, 20, 43200},
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

func Test_callAuction(t *testing.T) {
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
		{"callAuction test1", args{"cu1906", 40000}, 43500, 75, 15},
	}
	cleanupOrderBook("cu1906")
	buildOrBook(orders1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := callAuction(tt.args.sym, tt.args.pclose)
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

func Test_callAuctionNew(t *testing.T) {
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
		{"callAuctionNew test1", args{"cu1906", 40000}, 43500, 75, 15},
	}
	cleanupOrderBook("cu1906")
	buildOrBook(orders1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLast, gotMaxVol, gotVolRemain := callAuctionNew(tt.args.sym, tt.args.pclose)
			if gotLast != tt.wantLast {
				t.Errorf("callAuctionNew() gotLast = %v, want %v", gotLast, tt.wantLast)
			}
			if gotMaxVol != tt.wantMaxVol {
				t.Errorf("callAuctionNew() gotMaxVol = %v, want %v", gotMaxVol, tt.wantMaxVol)
			}
			if gotVolRemain != tt.wantVolRemain {
				t.Errorf("callAuctionNew() gotVolRemain = %v, want %v", gotVolRemain, tt.wantVolRemain)
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
	for i := 0; i < 2e5; i++ {
		price := rand.Intn(2000)*10 + pclose - 10000
		vol := rand.Intn(100) + 1
		SendOrder("cu1908", (i&1) != 0, vol, price)
	}
	// build cu1908 orderBook
	if ob, ok := simOrderBook["cu1908"]; ok {
		log.Infof("New orderBook bids: %d, asks: %d", ob.bids.Len(), ob.asks.Len())
	}
}

func Benchmark_callAuction(b *testing.B) {
	b.StopTimer()
	logging.SetLevel(logging.WARNING, "go-auction")
	buildBenchOrderBook()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := callAuction("cu1908", pclose)
		if i == 0 {
			b.Logf("callAuction price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}

func Benchmark_callAuctionNew(b *testing.B) {
	b.StopTimer()
	logging.SetLevel(logging.WARNING, "go-auction")
	buildBenchOrderBook()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		gotLast, vol, gotVolRemain := callAuctionNew("cu1908", pclose)
		if i == 0 {
			b.Logf("callAuction price:%d, volume:%d, remainVol:%d", gotLast, vol, gotVolRemain)
		}
	}
}
