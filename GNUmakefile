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

all: bin/auction
	@[ -d bin ] || exit

bin/auction:	cmd/auction/main.go
	@[ -d bin ] || mkdir bin
	@go build -o $@ cmd/auction/main.go
	@strip $@ || echo "auction OK"

clean:

distclean: clean
	@rm -rf bin
