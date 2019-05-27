package main

import (
	"auction"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/kjx98/golib/to"
	"github.com/op/go-logging"
)

var (
	orderFile   string
	longFile    string
	shortFile   string
	count       int
	algo        int
	verbose     bool
	testTrading bool
)

var log = logging.MustGetLogger("auction")
var pclose = 50000

const (
	instr = "cu1908"
)

func buildOrderBook(bTrading bool) {
	tt := time.Now()
	rand.Seed(tt.Unix())
	for i := 0; i < count; i++ {
		price := rand.Intn(20000) + pclose - 10000
		vol := rand.Intn(100) + 1
		auction.SendOrder(instr, (price&1) != 0, vol, price)
	}
	// build cu1908 orderBook
	et := time.Now()
	du := et.Sub(tt)
	if bTrading {
		log.Infof("Feed/Trade rand %d orders cost %.3f ms, %.2f O/s", count,
			du.Seconds()*1000.0, float64(count)/du.Seconds())
	} else {
		log.Infof("Build rand %d orders cost %.3f ms, %.2f O/s", count,
			du.Seconds()*1000.0, float64(count)/du.Seconds())
	}
}

func loadSideOrders(fileN string, isBuy bool) (cnt int) {
	if fd, err := os.Open(fileN); err != nil {
		log.Info("open", fileN, " error:", err)
	} else {
		defer fd.Close()
		rd := csv.NewReader(fd)
		tt := time.Now()
		lineCnt := 0
		if lines, err := rd.ReadAll(); err == nil {
			lineCnt = len(lines)
			prices := make([]int, lineCnt)
			vols := make([]int, lineCnt)
			for _, line := range lines {
				if len(line) < 3 {
					continue
				}
				prices[cnt] = to.Int(line[1])
				vols[cnt] = to.Int(line[2])
				cnt++
			}
			du := time.Now().Sub(tt)
			log.Infof("Read %s orders cost %.3f seconds", fileN, du.Seconds())
			tt = time.Now()
			for i := 0; i < cnt; i++ {
				vol := vols[i]
				pr := prices[i]
				auction.SendOrder(instr, isBuy, vol, pr)
			}
			du = time.Now().Sub(tt)
			log.Infof("Insert %d orders cost %.3f seconds", cnt, du.Seconds())
		}
	}
	return
}

func main() {
	flag.StringVar(&orderFile, "order", "", "csv format orders")
	flag.StringVar(&longFile, "long", "", "csv format long orders")
	flag.StringVar(&shortFile, "short", "", "csv format short orders")
	flag.IntVar(&count, "count", 2000000, "orders count")
	flag.IntVar(&algo, "algo", 1, "Call Auction Algorithm")
	flag.BoolVar(&verbose, "v", false, "verbose log")
	flag.BoolVar(&testTrading, "t", false, "test continuous trading")
	if !verbose {
		logging.SetLevel(logging.WARNING, "go-auction")
	}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: auction [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if orderFile != "" {
		if fd, err := os.Open(orderFile); err != nil {
			rcnt := 0
			tt := time.Now()
			rd := csv.NewReader(fd)
			if lines, err := rd.ReadAll(); err == nil {
				for _, line := range lines {
					var isBuy bool
					if len(line) < 2 {
						continue
					}
					pr := to.Int(line[0])
					vol := to.Int(line[1])
					if vol >= 0 {
						isBuy = true
					} else {
						vol = -vol
						isBuy = false
					}
					auction.SendOrder(instr, isBuy, vol, pr)
					rcnt++
				}
			}
			fd.Close()
			et := time.Now()
			du := et.Sub(tt)
			log.Infof("Load %d orders cost %.3f seconds, %.2f O/s", rcnt, du.Seconds(),
				float64(rcnt)/du.Seconds())
		} else {
			buildOrderBook(false)
		}
	} else if longFile != "" && shortFile != "" {
		tt := time.Now()
		longCnt := loadSideOrders(longFile, true)
		shortCnt := loadSideOrders(shortFile, false)
		et := time.Now()
		du := et.Sub(tt)
		log.Infof("Load %d long orders %d short orders cost %.3f seconds, %.2f O/s",
			longCnt, shortCnt, du.Seconds(), float64(longCnt+shortCnt)/du.Seconds())
	} else {
		buildOrderBook(false)
	}
	bLen, aLen := auction.OrderBookLen(instr)
	//fmt.Printf("集合竞价前报单簿, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
	fmt.Printf("Before auction, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
	tt := time.Now()
	var last, volume, remain int
	switch algo {
	case 0:
		last, volume, remain = auction.MatchCrossFill(instr, pclose)
	default:
		algo = 1
		fallthrough
	case 1:
		last, volume, remain = auction.MatchCross(instr, pclose)
	case 2:
		last, volume, remain = auction.MatchCrossOld(instr, pclose)
	case 3:
		bids, asks := auction.BuildOrBk(instr)
		di := time.Now().Sub(tt)
		fmt.Printf("Build bids, asks cost %.3f ms\n", di.Seconds()*1000)
		tt = time.Now()
		last, volume, remain = auction.CallAuction(bids, asks, pclose)
	}
	du := time.Now().Sub(tt)
	fmt.Printf("Auction Algo %d match %d orders cost %.3f ms, %.2f Ops\n",
		algo, count, du.Seconds()*1000.0, float64(count)/du.Seconds())
	fmt.Printf("CallAuction Price: %d, Volume: %d, Remain Volume: %d\n",
		last, volume, remain)
	if algo > 0 {
		tt = time.Now()
		auction.MatchOrder(instr, true, last, volume)
		auction.MatchOrder(instr, false, last, volume)
		du = time.Now().Sub(tt)
		//fmt.Printf("生成成交单耗时: %.3f ms\n", du.Seconds()*1000.0)
		fmt.Printf("Build deal reports cost: %.3f ms\n", du.Seconds()*1000.0)
	}

	bLen, aLen = auction.OrderBookLen(instr)
	//fmt.Printf("集合竞价后报单簿, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
	fmt.Printf("After auction, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
	if testTrading {
		auction.MarketStart(false)
		buildOrderBook(true)
		bLen, aLen = auction.OrderBookLen(instr)
		//fmt.Printf("连续交易后报单簿, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
		fmt.Printf("Trading continuous, bid QLen: %d, ask QLen: %d\n", bLen, aLen)
		cnt := auction.DealCount()
		//fmt.Printf("连续交易成交笔数: %d\n", cnt)
		fmt.Printf("matchs in continuous trading: %d\n", cnt)
	}
}

//  `%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`
func init() {
	var format = logging.MustStringFormatter(
		`%{color}%{time:01-02 15:04:05}  ▶ %{level:.4s} %{color:reset} %{message}`,
	)

	if runtime.GOARCH != "amd64" {
		format = logging.MustStringFormatter(
			`%{time:01-02 15:04:05} %{level:.4s} %{message}`,
		)
	}
	logback := logging.NewLogBackend(os.Stderr, "", 0)
	logfmt := logging.NewBackendFormatter(logback, format)
	logging.SetBackend(logfmt)
}
