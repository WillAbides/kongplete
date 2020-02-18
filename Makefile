GOCMD=go
GOBUILD=$(GOCMD) build

bin/bindown:
	./script/bootstrap-bindown.sh -b bin

bin/gobin: bin/bindown
	bin/bindown download $@

bin/golangci-lint: bin/bindown
	bin/bindown download $@

bin/goreadme: bin/gobin
	GOBIN=${CURDIR}/bin \
	bin/gobin github.com/posener/goreadme/cmd/goreadme

.PHONY: clean
clean:
	rm -rf ./bin
