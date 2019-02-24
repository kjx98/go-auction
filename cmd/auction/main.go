package main

import (
	"auction"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/kjx98/golib/to"

	"github.com/op/go-logging"
)

var orderFile = flag.String("order", "", "csv format orders")
var count = flag.Int("count", 1000000, "orders count")
var algo = flag.Int("algo", 1, "Call Auction Algorithm")
var log = logging.MustGetLogger("auction")
var pclose = 50000

const (
	instr = "cu1908"
)

func buildOrderBook() {
	tt := time.Now()
	rand.Seed(tt.Unix())
	for i := 0; i < *count; i++ {
		price := rand.Intn(2000)*10 + pclose - 10000
		vol := rand.Intn(100) + 1
		auction.SendOrder(instr, (i&1) != 0, vol, price)
	}
	// build cu1908 orderBook
	et := time.Now()
	du := et.Sub(tt)
	log.Infof("Build rand %d orders cost %.3f seconds, %g Ops", *count, du.Seconds(),
		float64(*count)/du.Seconds())
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: auction [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if *orderFile != "" {
		if fd, err := os.Open(*orderFile); err != nil {
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
			log.Infof("Load %d orders cost %.3f seconds, %g Ops", rcnt, du.Seconds(),
				float64(rcnt)/du.Seconds())
		} else {
			buildOrderBook()
		}
	} else {
		buildOrderBook()
	}
	tt := time.Now()
	var last, volume, remain int
	switch *algo {
	case 1:
		last, volume, remain = auction.MatchCross(instr, pclose)
	case 2:
		last, volume, remain = auction.MatchCrossOld(instr, pclose)
	case 3:
		last, volume, remain = auction.CallAuction(instr, pclose)
	}
	et := time.Now()
	du := et.Sub(tt)
	fmt.Printf("Auction Algo %d match %d orders cost %.3f seconds, %f Ops\n",
		*algo, *count, du.Seconds(), float64(*count)/du.Seconds())
	fmt.Printf("CallAuction Price: %d, Volume: %d, Remain Volume: %d\n",
		last, volume, remain)
}
