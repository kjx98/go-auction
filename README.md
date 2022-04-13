# go-auction
prototype of Match Engine in golang
try avl tree and rb tree for orderbook

Benchmark Cross/Continue match (avl tree for orderBook)
Cross for 2 million orders, buy/sell half/half
<pre>
goos: linux
goarch: amd64
pkg: github.com/kjx98/go-auction
cpu: Intel(R) Core(TM) i5-4200U CPU @ 1.60GHz
BenchmarkMatchCrossOld      	       6	 176611713 ns/op
BenchmarkMatchCross         	      12	  91508011 ns/op
BenchmarkMatchTradeContinue 	 1464553	       739.4 ns/op
PASS
</pre>
