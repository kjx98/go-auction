#
#	Makefile for hookAPI
#
# switches:
#	define the ones you want in the CFLAGS definition...
#
#	TRACE		- turn on tracing/debugging code
#
#
#
#

# Version for distribution
VER=1_0r1
GOPATH=$(shell go env GOPATH):$(PWD)

export GOPATH
MAKEFILE=GNUmakefile

# We Use Compact Memory Model

all: bin/auction bin/auction.exe
	@[ -d bin ] || exit

bin/auction:	cmd/auction/main.go
	@[ -d bin ] || mkdir bin
	@go build -o $@ cmd/auction/main.go
	@strip $@ || echo "auction OK"

bin/auction.exe:	cmd/auction/main.go
	@[ -d bin ] || mkdir bin
	GOOS=windows GOARCH=amd64 go build -o $@ cmd/auction/main.go
	#x86_64-w64-mingw32-strip $@

dtest: bin/auction
	@for a in 0 1 2 3; do \
	  GOGC=400 bin/auction -algo=$$a -long Data/long.txt -short Data/short.txt; \
	done

test:
	@(cd src/auction;go test )

bench:
	@(cd src/auction;GOGC=400 go test -bench=Match)

rbtest:
	@(cd src/auction;go test -tags rbtree)

rbbench:
	@(cd src/auction;GOGC=400 go test -tags rbtree -bench=Match)

clean:

distclean: clean
	@rm -rf bin
